package router

import (
	"StarDreamerCyberNook/api"
	"StarDreamerCyberNook/middleware"

	"github.com/gin-gonic/gin"
)

func UserFollowRouter(r *gin.RouterGroup) {
	app := api.App.FollowApi
	r.GET("/user/follow/list", app.FollowUserListView)                            //关注列表(不传userID查自己,传了查指定用户)
	r.GET("/user/follower/list", app.FollowerListView)                            //粉丝列表(不传userID查自己,传了查指定用户)
	r.GET("/user/friend/list", middleware.AuthMiddleware, app.FriendUserListView) //我的好友列表(需登录)

	r.POST("/user/follow", middleware.AuthMiddleware, app.FollowUserView)            //关注用户
	r.POST("/user/follow/unfollow", middleware.AuthMiddleware, app.UnFollowUserView) //取消关注用户
}
