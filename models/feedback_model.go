// models/feedback_model.go
// 用户反馈模型:收集用户提交的问题/建议,支持管理员回复与状态流转
package models

type FeedbackType int8

// 反馈类型
const (
	FeedbackTypeOther   FeedbackType = 0 // 其他
	FeedbackTypeFeature FeedbackType = 1 // 功能建议
	FeedbackTypeBug     FeedbackType = 2 // 问题反馈
	FeedbackTypeReport  FeedbackType = 3 // 内容举报
)

type FeedbackStatus int8

// 反馈处理状态
const (
	FeedbackStatusPending    FeedbackStatus = 0 // 待处理
	FeedbackStatusAccepted   FeedbackStatus = 1 // 已采纳未处理
	FeedbackStatusProcessing FeedbackStatus = 2 // 正在处理
	FeedbackStatusResolved   FeedbackStatus = 3 // 已处理
)

// FeedbackModel 用户反馈
type FeedbackModel struct {
	Model
	UserID      uint           `gorm:"index" json:"userID,omitempty"`     // 提交用户ID,0 表示未登录;匿名时不返回
	IsAnonymous bool           `gorm:"default:false" json:"isAnonymous"`  // 是否匿名提交(匿名时列表不展示提交者)
	Content     string         `gorm:"size:2048" json:"content"`          // 反馈内容
	Contact     string         `gorm:"size:128" json:"contact,omitempty"` // 联系方式(可选,便于回访;仅管理员接口返回)
	Type        FeedbackType   `gorm:"default:0" json:"type"`             // 反馈类型:0其他 1功能建议 2问题反馈 3内容举报
	Status      FeedbackStatus `gorm:"default:0" json:"status"`           // 处理状态:0待处理 1已采纳未处理 2正在处理 3已处理
	Reply       string         `gorm:"size:2048" json:"reply"`            // 管理员回复
	HandlerID   uint           `json:"handlerID,omitempty"`               // 处理人用户ID
}
