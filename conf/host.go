package conf

// Host 统一宿主配置:main 默认以宿主模式运行;
// 显式 enable: false 时退化为"仅博客"。
type Host struct {
	Enable *bool `yaml:"enable"` // 是否当宿主;缺省(nil)视为 true
}

// Enabled 返回是否启用宿主模式(未配置时默认 true)。
func (h Host) Enabled() bool {
	return h.Enable == nil || *h.Enable
}
