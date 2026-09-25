// models/moment_model.go
// 动态/日记模型定义
// 单模型用 type 区分动态与日记,用 visibility 控制 公开/仅好友/私密
package models

// MomentType 动态类型
type MomentType int8

const (
	MomentTypeMoment MomentType = 0 // 动态
	MomentTypeDiary  MomentType = 1 // 日记
)

func (t MomentType) String() string {
	switch t {
	case MomentTypeMoment:
		return "动态"
	case MomentTypeDiary:
		return "日记"
	default:
		return "未知"
	}
}

// MomentVisibility 可见性
type MomentVisibility int8

const (
	MomentVisibilityPublic  MomentVisibility = 0 // 公开
	MomentVisibilityFriends MomentVisibility = 1 // 仅好友
	MomentVisibilityPrivate MomentVisibility = 2 // 私密
)

// MomentModel 动态/日记模型
type MomentModel struct {
	Model
	UserID       uint             `gorm:"index;index:idx_moment_user_feed,priority:1" json:"userID"`                                          // 作者用户ID
	UserModel    UserModel        `gorm:"foreignKey:UserID" json:"user"`                                                                      // 作者信息
	Type         MomentType       `gorm:"default:0" json:"type"`                                                                              // 0动态 1日记
	Visibility   MomentVisibility `gorm:"default:0;index:idx_moment_feed,priority:2;index:idx_moment_user_feed,priority:3" json:"visibility"` // 0公开 1仅好友 2私密
	Content      string           `gorm:"size:16384" json:"content"`                                                                          // 正文
	Images       []string         `gorm:"type:text;serializer:json" json:"images"`                                                            // 图片URL列表,JSON序列化
	Mood         string           `gorm:"size:32" json:"mood"`                                                                                // 心情
	Location     string           `gorm:"size:64" json:"location"`                                                                            // 位置
	LikeCount    int              `json:"likeCount"`                                                                                          // 点赞数
	CommentCount int              `json:"commentCount"`                                                                                       // 评论数
	RepostCount  int              `json:"repostCount"`                                                                                        // 转发数
	RepostFromID *uint            `gorm:"index" json:"repostFromID"`                                                                          // 转发来源动态ID
	RepostFrom   *MomentModel     `gorm:"foreignKey:RepostFromID" json:"repostFrom"`                                                          // 原动态,外键关联(转发时随响应返回,已裁剪作者隐私字段)
	Status       Status           `gorm:"index:idx_moment_feed,priority:1;index:idx_moment_user_feed,priority:2" json:"status"`               // 复用文章状态:0草稿 1审核中 2已发布 3已下线
}

// MomentDiggModel 动态点赞模型
type MomentDiggModel struct {
	Model
	UserID   uint `gorm:"uniqueIndex:idx_moment_digg" json:"userID"`         // 点赞用户ID
	MomentID uint `gorm:"uniqueIndex:idx_moment_digg;index" json:"momentID"` // 动态ID
}

// MomentCommentModel 动态评论模型
// 复用评论的路径方案(父路径 + 根评论ID),支持嵌套回复
type MomentCommentModel struct {
	Model
	MomentID     uint                `gorm:"index;index:idx_moment_comment_root,priority:1;index:idx_moment_comment_path,priority:1" json:"momentID"`            // 所属动态ID
	MomentModel  MomentModel         `gorm:"foreignKey:MomentID" json:"-"`                                                                                       // 所属动态
	UserID       uint                `json:"userID"`                                                                                                             // 评论用户ID
	UserModel    UserModel           `gorm:"foreignKey:UserID" json:"user"`                                                                                      // 评论用户信息
	Content      string              `gorm:"size:1024" json:"content"`                                                                                           // 评论内容
	RootParentID *uint               `gorm:"index:idx_moment_comment_root,priority:2" json:"rootParentID"`                                                       // 根评论ID
	RootParent   *MomentCommentModel `gorm:"foreignKey:RootParentID" json:"-"`                                                                                   // 根评论,外键关联
	ParentPath   string              `gorm:"type:varchar(512);index;index:idx_moment_comment_path,priority:2;CHARACTER SET ascii COLLATE ascii_bin" json:"path"` // 评论路径,定位层级
	DiggCount    int                 `json:"diggCount"`                                                                                                          // 点赞数
}

// MomentCommentDiggModel 动态评论点赞模型
type MomentCommentDiggModel struct {
	Model
	UserID          uint `gorm:"uniqueIndex:idx_moment_comment_digg" json:"userID"`                // 点赞用户ID
	MomentCommentID uint `gorm:"uniqueIndex:idx_moment_comment_digg;index" json:"momentCommentID"` // 动态评论ID
}
