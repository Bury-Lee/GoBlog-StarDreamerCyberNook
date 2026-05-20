package utils

import "strings"

// highlightKeyword 简单模拟关键词高亮（降级模式使用）
func HighlightKeyword(text, keyword string) string {
	if keyword == "" || text == "" {
		return text
	}
	// 不区分大小写替换
	replaced := strings.ReplaceAll(text, keyword, "<em>"+keyword+"</em>")
	// 如果原文本包含大写形式的关键词，也需要处理
	if keyword != strings.ToLower(keyword) {
		replaced = strings.ReplaceAll(replaced, strings.ToLower(keyword), "<em>"+strings.ToLower(keyword)+"</em>")
		replaced = strings.ReplaceAll(replaced, strings.ToUpper(keyword), "<em>"+strings.ToUpper(keyword)+"</em>")
	}
	return replaced
}
