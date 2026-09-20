package utils

import "regexp"

// highlightKeyword 简单模拟关键词高亮（降级模式使用）
// 说明:用正则一次替换,不区分大小写;避免多次ReplaceAll导致标签嵌套(如 <em><em>key</em></em>)
func HighlightKeyword(text, keyword string) string {
	if keyword == "" || text == "" {
		return text
	}
	re := regexp.MustCompile("(?i)" + regexp.QuoteMeta(keyword))
	return re.ReplaceAllString(text, "<em>$0</em>")
}
