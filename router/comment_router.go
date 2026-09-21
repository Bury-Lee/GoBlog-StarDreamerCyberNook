package router

import (
	"StarDreamerCyberNook/api"
	"StarDreamerCyberNook/middleware"

	"github.com/gin-gonic/gin"
)

func CommentRouter(r *gin.RouterGroup) {
	app := api.App.CommentApi
	r.POST("/comment", middleware.AuthMiddleware, app.CommentCreateView)       //已测试
	r.DELETE("/comment/:id", middleware.AuthMiddleware, app.CommentDeleteView) //已测试
	r.GET("/comment", app.CommentListlView)                                    //已测试
	r.GET("/commentChild", app.CommentChildListView)                           //已测试
	// r.GET("/comment/interaction", middleware.AuthMiddleware, app.CommentInteractionView) //当前用户对一组评论的点赞状态,鉴于可能会对数据库造成较大压力,暂不添加
	r.POST("/comment/digg/:id", middleware.AuthMiddleware, app.CommentDiggView)
	r.POST("/comment/at/:id", middleware.AuthMiddleware, app.AtOther)
}
