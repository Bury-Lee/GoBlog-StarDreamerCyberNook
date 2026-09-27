package utils

// Str 读取字符串配置。
func Str(m map[string]any, key, def string) string {
	if m != nil {
		if v, ok := m[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return def
}

// Int 读取整型配置。
func Int(m map[string]any, key string, def int) int {
	if m != nil {
		if v, ok := m[key]; ok {
			switch n := v.(type) {
			case int:
				return n
			case int64:
				return int(n)
			case float64:
				return int(n)
			}
		}
	}
	return def
}

// Float 读取浮点配置。
func Float(m map[string]any, key string, def float64) float64 {
	if m != nil {
		if v, ok := m[key]; ok {
			switch n := v.(type) {
			case float64:
				return n
			case int:
				return float64(n)
			case int64:
				return float64(n)
			}
		}
	}
	return def
}

// Bool 读取布尔配置。
func Bool(m map[string]any, key string, def bool) bool {
	if m != nil {
		if v, ok := m[key]; ok {
			if b, ok := v.(bool); ok {
				return b
			}
		}
	}
	return def
}
