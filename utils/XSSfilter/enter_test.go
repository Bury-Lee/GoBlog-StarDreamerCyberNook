package xss_filter

import (
	"strings"
	"testing"
)

// TestSanitize 验证富文本过滤:危险标签/属性/协议被移除,白名单标签保留
func TestSanitize(t *testing.T) {
	filter := NewXSSFilter()
	cases := []struct {
		name           string
		input          string
		wantContains   string
		wantNotContain string
	}{
		{"script标签被移除", `<script>alert(1)</script>正常内容`, "正常内容", "<script>"},
		{"事件属性被移除", `<img src="https://example.com/a.png" onerror="alert(1)">图片`, "图片", "onerror"},
		{"javascript协议被移除", `<a href="javascript:alert(1)">点我</a>`, "点我", "javascript:"},
		{"白名单标签保留", `<strong>加粗</strong>文本`, "<strong>", ""},
	}
	for _, tc := range cases {
		got := filter.Sanitize(tc.input)
		if tc.wantContains != "" && !strings.Contains(got, tc.wantContains) {
			t.Errorf("%s: Sanitize(%q) = %q, 期望包含 %q", tc.name, tc.input, got, tc.wantContains)
		}
		if tc.wantNotContain != "" && strings.Contains(got, tc.wantNotContain) {
			t.Errorf("%s: Sanitize(%q) = %q, 不应包含 %q", tc.name, tc.input, got, tc.wantNotContain)
		}
	}
}

// TestSanitizeText 验证纯文本过滤:所有HTML标签被移除
func TestSanitizeText(t *testing.T) {
	got := SanitizeText("<b>昵称</b><script>alert(1)</script>")
	if strings.Contains(got, "<") || !strings.Contains(got, "昵称") {
		t.Errorf("SanitizeText 结果不符合预期: %q", got)
	}
}
