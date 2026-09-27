// Package blog 将博客主程序(HTTP 服务 + 定时任务)的启动/停止逻辑抽取为
// 可被宿主装载的运行时,使"博客 + 各微服务组件"能在同一个进程里按蓝图拉起。
package blog

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"StarDreamerCyberNook/core"
	"StarDreamerCyberNook/global"
	router "StarDreamerCyberNook/router"
	cron_service "StarDreamerCyberNook/service/cron"

	"github.com/sirupsen/logrus"
)

// App 是运行中的博客实例。
type App struct {
	server *http.Server
	cron   bool //是否开启定时任务
}

// Run 按 settingPath 配置路径完成初始化并启动 HTTP 服务(非阻塞;由 Stop 关闭)。
func Run(ctx context.Context, settingPath string) (*App, error) {
	global.Config = core.ReadConf(settingPath) //读取配置文件

	core.InitLogrus()                                             //初始化日志
	global.IPsearcher = core.InitIPDB()                           //初始化ip地址库
	global.DB = core.InitDB()                                     //初始化数据库
	global.RedisTimeCache, global.RedisHotPool = core.InitRedis() //初始化redis
	core.InitAIPrompt()                                           //定制看板娘提示词(模型调用已剥离到 ai 服务)
	if global.Config.ObjectStorage.Enable {
		logrus.Infof("对象存储已启用,存储桶:%s", global.Config.ObjectStorage.Bucket)
		global.StorageClient = core.InitClient()
	} else {
		logrus.Info("对象存储未启用,使用本地存储")
	}

	//debug模式下打印配置
	if global.Config.System.RunMode == "debug" {
		if configDebug, err := json.MarshalIndent(global.Config, "", "  "); err == nil {
			logrus.Debug(string(configDebug))
		}
	}

	if global.Config.System.Cron { //分布式环境建议只在一个实例启用定时任务
		go cron_service.Cron()
	}

	r := router.InitRouter() //注册路由
	server := core.InitServer(r)
	go func() { //非阻塞并由 Stop 关闭
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.Errorf("博客 HTTP 服务启动失败: %v", err)
		}
	}()
	return &App{server: server, cron: global.Config.System.Cron}, nil
}

// Stop 优雅关闭 HTTP 服务与定时任务。
func (a *App) Stop(ctx context.Context) error {
	if a.cron {
		cron_service.Stop()
	}
	if a.server == nil {
		return nil
	}
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return a.server.Shutdown(c)
}
