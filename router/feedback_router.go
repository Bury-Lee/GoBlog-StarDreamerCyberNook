package router

import (
	"StarDreamerCyberNook/api"
	"StarDreamerCyberNook/middleware"

	"github.com/gin-gonic/gin"
)

func FeedbackRouter(r *gin.RouterGroup) {
	app := api.App.FeedbackApi
	r.POST("/feedback", app.FeedbackCreateView)                                //提交反馈(登录可选/可匿名)
	r.GET("/feedback", app.FeedbackWallView)                                   //反馈墙(全站公开)
	r.PUT("/feedback/:id", middleware.AdminMiddleware, app.FeedbackHandleView) //处理反馈(管理员,增量更新)
}
