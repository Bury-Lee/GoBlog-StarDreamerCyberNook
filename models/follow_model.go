package models

import (
	"errors"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UserFollowModel struct { //以后改名为follow表
	Model
	UserID         uint      `json:"userID"` // 用户id
	UserModel      UserModel `gorm:"foreignKey:UserID" json:"-"`
	FocusUserID    uint      `json:"focusUserID"` // 关注的其他用户的ID
	FocusUserModel UserModel `gorm:"foreignKey:FocusUserID" json:"-"`
	Friend         bool      `json:"friend"` // 是否是好友,初始创建时一定得是false
}

// 注:当两个用户同时关注对方时就会导致双方只能建立关注关系而无法成为好友关系,这是高并发下出现的问题
// 因为此时对方的记录还未写入,但是检查已经执行,执行之后的结构就是没有对方的记录,所以只能建立关注关系
// 而且有死锁的风险
// AfterCreate 在创建关注记录后，检查是否形成双向关注，并更新好友状态
// 说明:同步执行。原实现在goroutine外defer cancel(),context会被立刻取消,
// 导致互关判定几乎必然失败、好友关系永远建立不起来,故改为在创建事务内同步完成
func (self *UserFollowModel) AfterCreate(DB *gorm.DB) error {
	if self.UserID == 0 || self.FocusUserID == 0 {
		return nil
	}

	// 查询是否存在反向关注，即对方也关注了自己
	var oppositeRelation UserFollowModel
	err := DB.Where(&UserFollowModel{
		UserID:      self.FocusUserID, // 对方的UserID
		FocusUserID: self.UserID,      // 对方关注的是我
	}).First(&oppositeRelation).Error
	if err != nil {
		// 如果没找到反向关注记录，则说明还不是好友，仅记录日志并返回
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.Errorf("查询反向关注记录失败: %v", err)
		}
		return nil
	}

	// 找到了反向关注，说明已经形成互相关注，需要更新双方的 friend 字段为 true
	// 启动一个事务来保证两个更新操作的原子性
	err = DB.Transaction(func(tx *gorm.DB) error {
		// 更新当前记录（self）的 friend 状态
		if err := tx.Model(self).Update("friend", true).Error; err != nil {
			return err // 返回错误以回滚事务
		}
		// 更新对方记录（oppositeRelation）的 friend 状态
		if err := tx.Model(&oppositeRelation).Update("friend", true).Error; err != nil {
			return err // 返回错误以回滚事务
		}
		return nil // 事务成功提交
	})
	if err != nil {
		// 如果事务失败，记录错误
		logrus.Errorf("好友关系更新失败: %v", err)
	}
	return nil
}

// BeforeDelete 在删除关注记录前，解除可能存在的好友关系
func (self *UserFollowModel) BeforeDelete(DB *gorm.DB) error {
	if self.UserID == 0 || self.FocusUserID == 0 {
		return nil
	}

	// 查询是否存在反向关注，即对方也关注了自己
	var oppositeRelation UserFollowModel
	err := DB.Where(&UserFollowModel{
		UserID:      self.FocusUserID, // 对方的UserID
		FocusUserID: self.UserID,      // 对方关注的是我
	}).First(&oppositeRelation).Error
	if err != nil {
		// 如果没找到反向关注记录，则说明当前删除操作不会破坏好友关系，仅记录日志并返回
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.Errorf("查询反向关注关系时出错: %v", err)
		}
		return nil
	}

	// 找到了反向关注，说明删除当前记录将破坏互相关注，需要解除双方的好友关系。
	// 注意：在此钩子中，我们只负责将对方记录的 friend 状态置为 false。
	// 当前记录（self）即将被删除，其 friend 状态无需再更新。
	if err := DB.Model(&oppositeRelation).Update("friend", false).Error; err != nil {
		// 如果更新失败，记录错误
		logrus.Errorf("解除好友关系失败 (user %d -> user %d): %v", oppositeRelation.UserID, oppositeRelation.FocusUserID, err)
	}
	return nil
}
