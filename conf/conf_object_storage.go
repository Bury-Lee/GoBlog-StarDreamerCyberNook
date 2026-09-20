package conf

// ObjectStorage 对象存储配置结构体
type ObjectStorage struct {
	Enable    bool   `yaml:"enable" json:"enable"`       // 是否启用
	AccessKey string `yaml:"accessKey" json:"accessKey"` // 访问密钥
	SecretKey string `yaml:"secretKey" json:"secretKey"` // 密钥
	Bucket    string `yaml:"bucket" json:"bucket"`       // 存储桶
	Host      string `yaml:"host" json:"host"`           // S3服务地址
	Uri       string `yaml:"uri" json:"uri"`             // URI
	Region    string `yaml:"region" json:"region"`       // 区域
	Prefix    string `yaml:"prefix" json:"prefix"`       // 前缀
	Size      uint   `yaml:"size" json:"size"`           // 文件大小限制(字节)
}
