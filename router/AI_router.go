package router

import (
	"StarDreamerCyberNook/api"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/middleware"

	"github.com/gin-gonic/gin"
)

func AIRouter(r *gin.RouterGroup) {
	//AI总开关或对话开关未开启时,不注册路由
	if !global.Config.AI.Enable || !global.Config.AI.ChatEnable {
		return
	}
	app := api.App.AIApi
	r.POST("/chat", middleware.AuthMiddleware, app.Chat)
}
