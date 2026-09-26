package router

import (
	"StarDreamerCyberNook/api"
	"StarDreamerCyberNook/middleware"

	"github.com/gin-gonic/gin"
)

func MomentRouter(r *gin.RouterGroup) {
	app := api.App.MomentApi
	//动态主体
	r.GET("/moment", app.MomentListView)                                                   //动态列表(隐私过滤)
	r.POST("/moment", middleware.AuthMiddleware, app.MomentCreateView)                     //发布动态/日记
	r.GET("/moment/interaction/:id", middleware.AuthMiddleware, app.MomentInteractionView) //当前用户点赞状态
	r.POST("/moment/digg/:id", middleware.AuthMiddleware, app.MomentDiggView)              //点赞/取消
	r.POST("/moment/repost/:id", middleware.AuthMiddleware, app.MomentRepostView)          //转发
	r.GET("/moment/:id", app.MomentDetailView)                                             //动态详情
	r.PUT("/moment", middleware.AuthMiddleware, app.MomentUpdateView)                      //更新(本人)
	r.DELETE("/moment/:id", middleware.AuthMiddleware, app.MomentRemoveView)               //删除(本人/管理员)
	//动态评论
	r.POST("/moment/comment", middleware.AuthMiddleware, app.MomentCommentCreateView)        //发表评论/回复
	r.GET("/moment/comment", app.MomentCommentListView)                                      //一级评论列表
	r.GET("/moment/comment/child", app.MomentCommentChildListView)                           //子评论列表
	r.DELETE("/moment/comment/:id", middleware.AuthMiddleware, app.MomentCommentDeleteView)  //删除评论
	r.POST("/moment/comment/digg/:id", middleware.AuthMiddleware, app.MomentCommentDiggView) //评论点赞
}
