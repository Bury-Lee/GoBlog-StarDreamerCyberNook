package cron_service

import (
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/service/ai_service"
	"StarDreamerCyberNook/service/review_service"
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	aiReviewBatchSize = 5               //每批审核的文章数
	aiReviewSleep     = 3 * time.Second //每批之间的休眠,降低AI服务与数据库压力
)

// aiReviewIng 避免上一次自动审核未结束就再次触发
var aiReviewIng sync.Mutex

// SyncAIReview 定时任务:将处于"审核中"的文章全部交给AI审核
// 说明:由配置 ai.auto_review 控制,AI服务不可用时直接跳过本次,待下次任务重试
func SyncAIReview() {
	if !global.Config.AI.AutoReview {
		return
	}
	// 尝试加锁，避免 cron 并发执行
	if !aiReviewIng.TryLock() {
		logrus.Info("自动AI审核任务正在执行,跳过本次任务")
		return
	}
	defer aiReviewIng.Unlock()

	//多实例互斥:与其它定时任务共用同一把分布式锁
	ctx := context.Background()
	if global.RedisTimeCache.Get(ctx, "cron_lock").Val() != global.Config.System.Addr() {
		logrus.Info("定时任务锁已被占用，跳过本次自动AI审核")
		return
	}

	//没有待审核文章时直接结束,避免无意义的AI探活请求
	hasPending, err := review_service.HasPendingArticles()
	if err != nil {
		logrus.Errorf("自动AI审核查询待审核文章失败: %v", err)
		return
	}
	if !hasPending {
		logrus.Info("没有需要AI审核的文章")
		return
	}

	//AI服务可用性探测:不可用时直接跳过,避免无意义的批量请求
	if ok, err := ai_service.Available(); !ok {
		logrus.Warnf("AI服务不可用,跳过本次自动审核: %v", err)
		return
	}

	total, success, failed, err := review_service.ReviewAllPendingArticles(aiReviewBatchSize, aiReviewSleep)
	if err != nil {
		logrus.Errorf("自动AI审核执行失败: %v", err)
		return
	}
	if total == 0 {
		logrus.Info("没有需要AI审核的文章")
		return
	}
	logrus.Infof("自动AI审核完成,本次处理 %d 篇,成功 %d 篇,失败 %d 篇", total, success, failed)
}
