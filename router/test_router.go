package router

import (
	"StarDreamerCyberNook/api"

	"github.com/gin-gonic/gin"
)

func TestRouter(nr *gin.RouterGroup) {
	//测试路由
	nr = nr.Group("/t")
	app := api.App.TestApi
	nr.GET("/test", app.TestView)
	nr.POST("/print", app.Print)
	nr.GET("/config", app.GetConfig)  //测试:查看运行时配置
	nr.POST("/config", app.SetConfig) //测试:修改运行时配置
}
