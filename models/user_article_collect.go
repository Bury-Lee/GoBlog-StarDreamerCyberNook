// models/user_article_collect_model.go
package models

import (
	"StarDreamerCyberNook/service/redis_service/redis_count"
	"time"

	"gorm.io/gorm"
)

type UserArticleCollectModel struct { //用户收藏文章表.目前只能收藏到一个收藏夹，后续可能会改为多对多关系
	UserID       uint         `gorm:"uniqueIndex:idx_user_article_collect,priority:1" json:"userID"`    // 用户id
	UserModel    UserModel    `gorm:"foreignKey:UserID" json:"-"`                                       // 收藏者
	ArticleID    uint         `gorm:"uniqueIndex:idx_user_article_collect,priority:2" json:"articleID"` // 文章id
	ArticleModel ArticleModel `gorm:"foreignKey:ArticleID" json:"-"`                                    // 被收藏的文章
	CollectID    uint         `gorm:"index:idx_user_article_collect_folder" json:"collectID"`           // 收藏夹的id
	CollectModel CollectModel `gorm:"foreignKey:CollectID" json:"collectModel"`                         // 属于哪一个收藏夹
	CreatedAt    time.Time    `json:"createdAt"`                                                        // 收藏的时间
}

// AfterCreate 创建时对应文章的收藏数加1
// 计数统一由模型钩子维护,业务代码不要再手动调用redis_count,否则会双写导致计数漂移
func (u *UserArticleCollectModel) AfterCreate(db *gorm.DB) error {
	if u.ArticleID == 0 {
		return nil
	}
	redis_count.SetCacheCollect(u.ArticleID, true)
	return nil
}

// BeforeDelete 删除时对应文章的收藏数减1
// 仅当删除的是真实记录(带主键)时生效;批量删除请使用SkipHooks并自行调整计数
func (u *UserArticleCollectModel) BeforeDelete(db *gorm.DB) error {
	if u.ArticleID == 0 {
		return nil
	}
	redis_count.SetCacheCollect(u.ArticleID, false)
	return nil
}
