package xss_filter

import (
	"regexp"

	"github.com/microcosm-cc/bluemonday"
)

type AdvancedXSSFilter struct {
	policy *bluemonday.Policy
}

func NewXSSFilter() *AdvancedXSSFilter {
	policy := bluemonday.NewPolicy()
	policy.AllowElements(
		"p", "br", "strong", "em",
		"h1", "h2", "h3", "h4",
		"ul", "ol", "li", "a",
		"img", "div", "span",
	)
	policy.AllowAttrs("href", "title").OnElements("a")
	policy.AllowAttrs("src", "alt", "title").OnElements("img")
	policy.AllowAttrs("class").Matching(regexp.MustCompile(`^[a-zA-Z0-9_\-\s]+$`)).OnElements("div", "span")
	policy.AllowStandardURLs()
	policy.AllowURLSchemes("http", "https", "mailto", "tel")
	policy.AllowDataURIImages()
	policy.RequireNoFollowOnLinks(true)
	policy.RequireNoReferrerOnLinks(true)
	policy.AddTargetBlankToFullyQualifiedLinks(true)

	return &AdvancedXSSFilter{policy: policy}
}

func (f *AdvancedXSSFilter) Sanitize(input string) string {
	if f == nil || f.policy == nil {
		return ""
	}
	return f.policy.Sanitize(input)
}

var strictPolicy = bluemonday.StrictPolicy()

// SanitizeText 清洗纯文本字段(标题/昵称/简介等),移除全部HTML标签
func SanitizeText(input string) string {
	return strictPolicy.Sanitize(input)
}
