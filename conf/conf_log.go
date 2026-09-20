package conf

type Log struct {
	App      string `yaml:"app"`       //来自哪个服务
	Dir      string `yaml:"dir"`       //日志目录
	LogLevel string `yaml:"log_level"` //记录的最低日志级别
}
