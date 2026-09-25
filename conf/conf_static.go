package conf

// Static 前端静态资源托管配置
type Static struct {
	Dir          string `yaml:"dir" json:"dir"`                   // 静态资源根目录
	WebPrefix    string `yaml:"webPrefix" json:"webPrefix"`       // 通用静态资源访问前缀,如 /web
	AssetsPrefix string `yaml:"assetsPrefix" json:"assetsPrefix"` // 构建产物(js/css)访问前缀,需与前端构建配置一致
	Favicon      string `yaml:"favicon" json:"favicon"`           // favicon 相对 Dir 的路径
	Index        string `yaml:"index" json:"index"`               // SPA 入口文件相对 Dir 的路径
}
