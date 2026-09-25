package ai_service

// CommentArticle 调用AI生成文章的质量评级与摘要,返回(评级, 摘要, 错误)
func CommentArticle(title, abstract, content string) (quality, summary string, err error) {
	summary, err = GenerateArticleAbstract(title, abstract, content)
	if err != nil {
		return "", "", err
	}
	quality, err = GenerateArticleQuality(title, abstract, content)
	if err != nil {
		return "", "", err
	}
	return quality, summary, nil
}
