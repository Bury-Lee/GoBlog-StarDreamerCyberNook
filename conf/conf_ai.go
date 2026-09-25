package conf

type AI struct { //AI模型配置
	Enable bool `yaml:"enable" json:"enable"` // 是否启用AI模型

	ChatEnable bool `yaml:"chat_enable" json:"chat_enable"` // 是否开放AI对话接口,单独控制,关闭后 /api/chat 不可用

	AutoReview bool `yaml:"auto_review" json:"auto_review"` // 是否启用定时任务自动AI审核(将待审核文章全部交给AI,AI服务不可用时自动跳过)

	AutoComment bool `yaml:"auto_comment" json:"auto_comment"` // 是否启用定时任务补全AI点评(为缺少点评记录的文章生成评级+摘要)

	Model       string  `yaml:"model" json:"model"`             // AI模型名称,为local时使用本地模型
	Temperature float32 `yaml:"temperature" json:"temperature"` // 温度参数，控制生成文本的随机性
	MaxTokens   int     `yaml:"max_tokens" json:"max_tokens"`   // 最大生成令牌数
	Host        string  `yaml:"host" json:"host"`               // 本地AI模型主机地址,默认http://localhost:1234/v1,当model为local时生效
	APIType     string  `yaml:"api_type" json:"api_type"`       // AI模型API类型,默认openai

	ApiKey   string `yaml:"ApiKey" json:"-"`          // AI模型密钥
	NickName string `yaml:"nickName" json:"nickName"` // AI模型昵称
	Avatar   string `yaml:"avatar" json:"avatar"`     // AI模型头像URL
	Platform string `yaml:"platform" json:"platform"` // AI模型平台TODO:后期适配
}
