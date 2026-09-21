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

const (
	cleanHistoryBatchSize = 200                //每批清理条数
	cleanHistorySleep     = 200 * time.Millisecond //每批之间的休眠,降低数据库压力
)

// SyncCleanHistory 清理超过30天的浏览记录
// 采用ID游标(id > cursor)分批推进,避免反复从表头扫描导致大表清理越来越慢
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

	var cursor uint   //游标:上一批处理到的最大ID
	total := 0        //本次清理总数
	batchIndex := 0   //批次序号
	logrus.Infof("开始清理超过一个月的浏览记录")

	for {
		if time.Now().After(deadline) {
			logrus.Warnf("清理浏览记录超时,提前结束,本次已清理 %d 条", total)
			return
		}

		//按主键游标顺序取一批过期记录,只查主键,避免加载整行数据
		var batch []models.UserArticleHistoryModel
		err := global.DB.
			Select("id").
			Where("id > ? and created_at < ?", cursor, expireTime).
			Order("id asc").
			Limit(cleanHistoryBatchSize).
			Find(&batch).Error
		if err != nil {
			logrus.Errorf("查询待清理浏览记录失败: %v", err)
			return
		}

		// 没有更多过期记录,说明已经清理完了
		if len(batch) == 0 {
			logrus.Infof("浏览记录清理完成,本次共清理 %d 条", total)
			return
		}

		ids := make([]uint, 0, len(batch))
		for _, item := range batch {
			ids = append(ids, item.ID)
		}

		//按主键批量删除,走主键索引
		tx := global.DB.Delete(&models.UserArticleHistoryModel{}, ids)
		if tx.Error != nil {
			logrus.Errorf("清理失败: %v", tx.Error)
			return
		}

		cursor = ids[len(ids)-1]
		total += int(tx.RowsAffected)
		batchIndex++
		if batchIndex%10 == 0 {
			logrus.Infof("浏览记录清理中,已处理 %d 条,当前游标 %d", total, cursor)
		}

		// 每批之间休眠，避免数据库压力过大
		time.Sleep(cleanHistorySleep)
	}
}
