package conf

import (
	"testing"

	"gopkg.in/yaml.v2"
)

// TestComponentConfigUnmarshal 验证 components 段同时支持 bool 简写与 {enabled, config} 完整写法。
func TestComponentConfigUnmarshal(t *testing.T) {
	var c struct {
		Components map[string]ComponentConfig `yaml:"components"`
	}
	yml := `
components:
  blog: true
  chat: false
  content:
    enabled: true
    config:
      dbStandalone: true
      driver: mysql
`
	if err := yaml.Unmarshal([]byte(yml), &c); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !c.Components["blog"].Enabled {
		t.Fatal("blog 应为启用")
	}
	if c.Components["chat"].Enabled {
		t.Fatal("chat 应为关闭")
	}
	content := c.Components["content"]
	if !content.Enabled {
		t.Fatal("content 应为启用")
	}
	if content.Config["dbStandalone"] != true {
		t.Fatalf("content.config.dbStandalone = %v, want true", content.Config["dbStandalone"])
	}
	if content.Config["driver"] != "mysql" {
		t.Fatalf("content.config.driver = %v, want mysql", content.Config["driver"])
	}
}
