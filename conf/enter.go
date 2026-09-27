package conf

const Version = "1.1.0"

// 记录全局变量的定义及其初始化
type Config struct {
	System        System          `yaml:"system"`
	Jwt           Jwt             `yaml:"jwt"`
	Log           Log             `yaml:"log"`
	ES            ES              `yaml:"es"`
	RedisStatic   Redis           `yaml:"redisStatic"`
	RedisDynamic  Redis           `yaml:"redisDynamic"`
	DBWrite       []DB            `yaml:"dbWrite"` // 写库列表
	DBRead        []DB            `yaml:"dbRead"`  // 读库列表
	Upload        UploadConfig    `yaml:"upload"`  //图片上传配置
	Email         Email           `yaml:"email"`
	AI            AI              `yaml:"ai"`
	ObjectStorage ObjectStorage   `yaml:"objectStorage"`
	Static        Static          `yaml:"static"`
	Site          Site            `yaml:"site"`
	QQ            QQ              `yaml:"qq" json:"qq"`
	Host          Host                       `yaml:"host"`       //统一宿主开关(是否当宿主)
	Components    map[string]ComponentConfig `yaml:"components"` //启用哪些微服务组件(缺省=全部启用)

	// Services 是微服务地址表:服务键 → 实例地址列表。
	// 例如: services: { ai: ["127.0.0.1:9210"], search: ["127.0.0.1:9220"] }
	// 环境变量 SVC_<KEY> 可选覆盖。
	Services map[string][]string `yaml:"services"`
}

// ComponentConfig 描述单个组件的启用开关与可选配置,兼容两种写法:
//
//	components:
//	  blog: true                       # 简写:仅声明启用
//	  content:                         # 完整:启用 + 组件配置
//	    enabled: true
//	    config:
//	      dbStandalone: true           # 该服务使用独立数据库
//	      driver: mysql
//	      dsn: "user:pass@tcp(127.0.0.1:3306)/blog"
type ComponentConfig struct {
	Enabled bool           `yaml:"enabled"`
	Config  map[string]any `yaml:"config"`
}

// UnmarshalYAML 兼容 bool 简写与 {enabled, config} 完整写法。
func (c *ComponentConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var b bool
	if err := unmarshal(&b); err == nil {
		c.Enabled = b
		return nil
	}
	type raw ComponentConfig // 避免递归调用本方法
	var r raw
	if err := unmarshal(&r); err != nil {
		return err
	}
	*c = ComponentConfig(r)
	return nil
}
