package cron_service

import (
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// 清理超过一个月的浏览记录
var cleanIng sync.Mutex

func SyncCleanHistory() {
	// 尝试加锁，避免 cron 并发执行
	if !cleanIng.TryLock() {
		return
	}
	defer cleanIng.Unlock()

	//多实例互斥:与计数同步任务共用同一把分布式锁
	ctx := context.Background()
	if global.RedisTimeCache.Get(ctx, "cron_lock").Val() != global.Config.System.Addr() {
		logrus.Info("清理任务锁已被占用,跳过本次清理")
		return
	}

	// 计算过期时间（30天前）
	expireTime := time.Now().AddDate(0, 0, -30)
	//设置最长执行时间,避免数据量过大时长时间占用数据库连接
	deadline := time.Now().Add(30 * time.Minute)

	for {
		if time.Now().After(deadline) {
			logrus.Warn("清理浏览记录超时,提前结束")
			return
		}
		logrus.Infof("开始清理超过一个月的浏览记录")

		// 执行删除（每次最多删除50条）
		tx := global.DB.
			Where("created_at < ?", expireTime).
			Limit(50).
			Delete(&models.UserArticleHistoryModel{})

		if tx.Error != nil {
			logrus.Errorf("清理失败: %v", tx.Error)
			return
		}

		affected := tx.RowsAffected

		// 如果本次一条都没删，说明已经清理完了
		if affected == 0 {
			logrus.Infof("浏览记录清理完成")
			return
		}

		// logrus.Infof("本次清理 %d 条记录", affected)

		// 每批之间休眠，避免数据库压力过大
		time.Sleep(5 * time.Second)
	}
}
