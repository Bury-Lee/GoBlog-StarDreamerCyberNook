// router/enter.go
package router

import (
	_ "embed"
	"os"
	"path/filepath"
	"strings"

	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/middleware"

	"github.com/gin-gonic/gin"
)

//go:embed 404page.html
var UnFoundPage string //这是404页面的HTML内容,用于自定义404页面,由于前端不想做,所以这里直接写在这里了

func InitRouter() *gin.Engine {
	gin.SetMode(global.Config.System.RunMode) //设置gin模式
	r := gin.Default()

	//注册静态资源与前端路由,作用相当于Nginx
	static := global.Config.Static
	var spaIndexFile string
	if static.Dir != "" {
		if static.Index != "" {
			spaIndexFile = filepath.Join(static.Dir, static.Index)
		}
		if static.WebPrefix != "" {
			r.Static(static.WebPrefix, static.Dir) //静态文件目录,可用于获取图片等
		}
		if static.AssetsPrefix != "" {
			r.Static(static.AssetsPrefix, filepath.Join(static.Dir, "assets")) //构建产物里的 js/css
		}
		if static.Favicon != "" {
			r.StaticFile("/favicon.svg", filepath.Join(static.Dir, static.Favicon))
		}
		if spaIndexFile != "" {
			r.StaticFile("/", spaIndexFile) //前端 SPA 直接由后端托管,http://host:8080/
		}
	}

	nr := r.Group("/api") //TODO:测试使用无前缀api,开发完成了要给app组加上/api的前缀,nr := r.Group("/api")这样

	if global.Config.System.RunMode == "debug" {
		//注册在引擎级别,保证 OPTIONS 预检请求(无匹配路由)也能通过跨域中间件
		r.Use(middleware.RequestLogMiddleware)
		r.Use(middleware.CORS)
		TestRouter(nr)
	}
	nr.Use(middleware.SecurityHeaders) //基础安全响应头
	nr.Use(middleware.LogMiddleware)
	nr.Use(middleware.ActLimitMiddleware) //防攻击中间件
	HearthRouter(nr)                      //心跳路由注册函数

	SiteRouter(nr) //已测试完毕
	LogRouter(nr)  //由于条件问题,待测

	if global.Config.ObjectStorage.Enable {
		OSSImageRouter(nr)
	} else {
		//如果对象存储未启用,则使用本地存储
		LocalImageRouter(nr) //已测试完毕
	}
	BannerRouter(nr)   //已测试完毕
	UserRouter(nr)     //已测试完毕
	CaptcharRouter(nr) //已测试完毕
	ArticleRouter(nr)  //已测试完毕
	CommentRouter(nr)  //已测试完毕
	MessageRouter(nr)
	ChatRouter(nr)
	AIRouter(nr) //已测试完毕
	FriendRouter(nr)
	UserFollowRouter(nr) //关注/粉丝/好友
	FeedbackRouter(nr)   //用户反馈

	r.NoRoute(func(ctx *gin.Context) { //没命中任何路由时的兜底
		path := ctx.Request.URL.Path
		//接口和后端静态资源不存在:返回内置的404页面
		if strings.HasPrefix(path, "/api") || (static.WebPrefix != "" && strings.HasPrefix(path, static.WebPrefix)) {
			ctx.Data(404, "text/html; charset=utf-8", []byte(UnFoundPage))
			return
		}
		//其余路径交给前端路由(history模式),刷新 /u/1 这类页面才不会404
		if spaIndexFile == "" {
			ctx.Data(404, "text/html; charset=utf-8", []byte(UnFoundPage))
			return
		}
		if _, err := os.Stat(spaIndexFile); err != nil {
			ctx.Data(404, "text/html; charset=utf-8", []byte(UnFoundPage))
			return
		}
		ctx.File(spaIndexFile)
	})

	return r
}
