package cron_service

import (
	"StarDreamerCyberNook/global"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

var crontab *cron.Cron

// Stop 停止定时任务,用于优雅退出
func Stop() {
	if crontab != nil {
		crontab.Stop()
	}
}

func Cron() {
	//抢锁执行

	timezone, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		logrus.Warnf("定时任务无法设置时区,已使用UTC时区:%v", err)
		crontab = cron.New(cron.WithSeconds(), cron.WithLocation(time.UTC))
	} else {
		crontab = cron.New(cron.WithSeconds(), cron.WithLocation(timezone))
	}

	//debug使用
	// crontab.AddFunc("*/10 * * * * *", SyncArticle)//debug:10秒一次
	// crontab.AddFunc("*/4 * * * * *", SyncArticle)
	// crontab.AddFunc("*/4 * * * * *", SyncComment)

	// 每10分钟，0秒触发
	crontab.AddFunc("0 */10 * * * *", GetLock) //10分钟的锁,每10分钟抢一次

	// 每10分钟
	crontab.AddFunc("0 1-59/10 * * * *", SyncArticle) //文章数据同步

	// 每10分钟
	crontab.AddFunc("0 6-59/10 * * * *", SyncComment) //评论数据同步

	if global.Config.System.ScheduledCleanup {
		crontab.AddFunc("0 */10 * * * *", SyncCleanHistory) //10分钟尝试清理一次浏览记录
	}

	//每分钟把待审核文章交给AI审核(由运行时配置 ai.auto_review 控制),错开文章/评论同步的秒点
	crontab.AddFunc("0 3-59/1 * * * *", SyncAIReview) //定时自动AI审核

	// 每天凌晨4:30对账一次,通过以关系表为基准修正文章计数的漂移
	// crontab.AddFunc("0 30 4 * * *", ReconcileArticleCount)

	crontab.Start()
}
