package cron_service

import (
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/ai_service"
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	aiCommentBatchSize              = 5               //单次任务最多补全的文章数
	aiCommentSleep                  = 3 * time.Second //每篇之间的休眠,降低AI服务与数据库压力
	maxConsecutiveAICommentFailures = 10              //连续失败达到该次数时提前结束,避免AI不可用时空转
)

// aiCommentIng 避免上一次自动点评未结束就再次触发
var aiCommentIng sync.Mutex

// SyncAIComment 定时任务:检查缺少AI点评记录的文章并补全(评级+摘要)
// 说明:由配置 ai.auto_comment 控制,AI服务不可用时直接跳过,待下次任务重试
func SyncAIComment() {
	if !global.Config.AI.Enable || !global.Config.AI.AutoComment {
		return
	}
	if !aiCommentIng.TryLock() {
		logrus.Info("自动AI点评任务正在执行,跳过本次任务")
		return
	}
	defer aiCommentIng.Unlock()

	//多实例互斥:与其它定时任务共用同一把分布式锁
	ctx := context.Background()
	if global.RedisTimeCache.Get(ctx, "cron_lock").Val() != global.Config.System.Addr() {
		logrus.Info("定时任务锁已被占用,跳过本次自动AI点评")
		return
	}

	//AI服务可用性探测:不可用时直接跳过,避免无意义的批量请求
	if ok, err := ai_service.Available(); !ok {
		logrus.Warnf("AI服务不可用,跳过本次自动点评: %v", err)
		return
	}

	total, success, failed := backfillAIComment(aiCommentBatchSize)
	if total == 0 {
		logrus.Info("没有需要补全AI点评的文章")
		return
	}
	logrus.Infof("自动AI点评完成,本次处理 %d 篇,成功 %d 篇,失败 %d 篇", total, success, failed)
}

// backfillAIComment 按ID游标补全缺少AI点评记录的文章
// 说明:通过 NOT EXISTS 过滤已有点评记录的文章,保证每篇文章只点评一次;
// AI调用失败时不写记录,留待下次任务重试
func backfillAIComment(batch int) (total, success, failed int) {
	var cursor uint
	consecutiveFailures := 0

	for i := 0; i < batch; i++ {
		var article models.ArticleModel
		err := global.DB.
			Where("id > ?", cursor).
			Where("NOT EXISTS (SELECT 1 FROM article_additions WHERE article_additions.article_id = article_models.id)").
			Order("id asc").
			First(&article).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			break
		}
		if err != nil {
			logrus.Errorf("查询待点评文章失败: %v", err)
			break
		}
		cursor = article.ID
		total++

		quality, summary, err := ai_service.CommentArticle(article.Title, article.Abstract, article.Content)
		if err != nil {
			failed++
			consecutiveFailures++
			logrus.Errorf("AI点评文章 %d 失败: %v", article.ID, err)
			time.Sleep(4 * time.Second)
			if consecutiveFailures >= maxConsecutiveAICommentFailures {
				logrus.Warnf("AI服务连续 %d 次不可用,提前结束本次自动点评", consecutiveFailures)
				break
			}
			continue
		}
		//以 article_id 为键 upsert 到文章扩展附录表(记录本次点评的AI模型名)
		if e := global.DB.Where(models.ArticleAddition{ArticleID: article.ID}).
			Assign(map[string]any{
				"ai_quality":  quality,
				"ai_abstract": summary,
				"ai_model":    global.Config.AI.Model,
			}).
			FirstOrCreate(&models.ArticleAddition{}).Error; e != nil {
			failed++
			logrus.Errorf("保存文章 %d 的AI点评失败: %v", article.ID, e)
			continue
		}
		success++
		consecutiveFailures = 0
		//清理详情缓存,让新补全的AI点评尽快对前端可见
		global.RedisHotPool.Del(context.Background(), "ArticleID"+strconv.FormatUint(uint64(article.ID), 10))
		time.Sleep(aiCommentSleep)
	}
	return total, success, failed
}
