package conf

type ES struct {
	Enabled  bool   `yaml:"enabled"`
	Url      string `yaml:"url"`
	UserName string `yaml:"username"`
	Password string `yaml:"password"`
}
