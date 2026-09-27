package core

import (
	"StarDreamerCyberNook/global"

	"github.com/sirupsen/logrus"
)

// InitAIPrompt 依据配置定制看板娘人格提示词。
// 说明:模型调用已剥离到 ai 微服务,博客不再持有模型客户端。
func InitAIPrompt() {
	if !global.Config.AI.Enable {
		logrus.Info("AI模型已禁用")
		return
	}

	if global.Config.AI.NickName != "" || global.Config.Site.Project.Title != "" {
		words := "你是" + global.Config.AI.NickName + "，" +
			global.Config.Site.Project.Title + " 网站的官方看板娘。" +
			"性格设定：活泼可爱、略带科技感、对用户友好。\n" +
			"回答要求：\n" +
			"- 简洁明了，控制在50字以内\n" +
			"- 使用中文回复\n" +
			"- 可适当使用颜文字或emoji增加亲和力\n" +
			"- 拒绝回答涉及敏感政治、违法犯罪、色情暴力等内容"

		global.SystemPromptMainSite = global.SystemPrompt(words)
	} else {
		logrus.Infof("未配置ai昵称和网站名称,已启用默认设置")
	}
}
