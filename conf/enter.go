package conf

const Version = "1.1.0"

// 记录全局变量的定义及其初始化
type Config struct {
	System        System        `yaml:"system"`
	Jwt           Jwt           `yaml:"jwt"`
	Log           Log           `yaml:"log"`
	ES            ES            `yaml:"es"`
	RedisStatic   Redis         `yaml:"redisStatic"`
	RedisDynamic  Redis         `yaml:"redisDynamic"`
	DBWrite       []DB          `yaml:"dbWrite"` // 写库列表
	DBRead        []DB          `yaml:"dbRead"`  // 读库列表
	Upload        UploadConfig  `yaml:"upload"` //图片上传配置
	Email         Email         `yaml:"email"`
	AI            AI            `yaml:"ai"`
	ObjectStorage ObjectStorage `yaml:"objectStorage"`
	Static        Static        `yaml:"static"`
	Site          Site          `yaml:"site"`
	QQ            QQ            `yaml:"qq" json:"qq"`
}
