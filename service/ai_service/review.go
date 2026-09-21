package ai_service

import (
	"StarDreamerCyberNook/global"
	"strings"
)

// AI审核判定结果
const (
	AIReviewPass   = "通过"
	AIReviewReject = "拒绝"
)

// ReviewArticle 调用AI对文章进行合规性审核
// 返回:"通过" / "拒绝" / 其它无法判定的原始回复
func ReviewArticle(title, abstract, content string) (string, error) {
	reply, err := CreateSingleReply(
		"文章标题:"+title+"\n文章摘要:"+abstract+"\n文章内容:"+content,
		global.SystemPromptArticleReview.String(),
	)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(reply), nil
}

// GenerateArticleAbstract 调用AI生成文章摘要
func GenerateArticleAbstract(title, abstract, content string) (string, error) {
	reply, err := CreateSingleReply(
		"文章标题:"+title+"\n文章摘要:"+abstract+"\n文章内容:"+content,
		global.SystemPromptArticleAbstract.String(),
	)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(reply), nil
}

// GenerateArticleQuality 调用AI生成文章质量评级
func GenerateArticleQuality(title, abstract, content string) (string, error) {
	reply, err := CreateSingleReply(
		"文章标题:"+title+"\n文章摘要:"+abstract+"\n文章内容:"+content,
		global.SystemPromptArticleAiQuality.String(),
	)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(reply), nil
}
