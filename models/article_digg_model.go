// models/article_digg_model.go
package models

import "time"

// ArticleDiggModel 点赞记录模型
type ArticleDiggModel struct { //UserID和ArticleID共同使用同一个复合唯一索引,保证同一用户对同一文章只能点赞一次
	Model
	UserID       uint         `gorm:"uniqueIndex:idx_article_digg_user_article,priority:1" json:"userID"`    // 用户ID
	ArticleID    uint         `gorm:"uniqueIndex:idx_article_digg_user_article,priority:2" json:"articleID"` // 文章ID
	UserModel    UserModel    `gorm:"foreignKey:UserID" json:"-"`                                            // 用户信息（关联）
	ArticleModel ArticleModel `gorm:"foreignKey:ArticleID" json:"-"`                                         // 文章信息（关联）
	CreatedAt    time.Time    `json:"createdAt"`                                                             // 点赞时间
}
