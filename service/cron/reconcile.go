package cron_service

import (
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/redis_service/redis_count"
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	reconcileBatchSize = 200
	reconcileSleep     = 200 * time.Millisecond
)

type articleCountRow struct {
	ArticleID uint
	Cnt       int
}

// ReconcileArticleCount 文章计数对账任务
// 说明:点赞/收藏/评论的计数是"先写Redis增量,再定时回写数据库",一旦进程崩溃或操作失败就可能漂移。
// 这里以关系表(article_digg_models / user_article_collect_models / comment_models)为基准反算真实值,
// 数据库应存的值 = 关系表实际值 - Redis中尚未同步的增量,
// 这样下一轮增量同步写入后,数据库值正好等于关系表实际值。
// 浏览数无法从关系表反算(浏览记录会被清理),因此不参与对账。
func ReconcileArticleCount() {
	//与增量同步任务互斥,避免对账读到同步中途的状态而误判
	if !articleCountSyncMutex.TryLock() {
		logrus.Info("文章计数同步任务正在执行,跳过本次对账")
		return
	}
	defer articleCountSyncMutex.Unlock()

	//多实例互斥:与其它定时任务共用同一把分布式锁
	ctx := context.Background()
	if global.RedisTimeCache.Get(ctx, "cron_lock").Val() != global.Config.System.Addr() {
		logrus.Info("对账任务锁已被占用,跳过本次对账")
		return
	}

	deadline := time.Now().Add(30 * time.Minute)
	var cursor uint
	scanned := 0
	fixed := 0

	for {
		if time.Now().After(deadline) {
			logrus.Warnf("文章计数对账超时,提前结束,已扫描 %d 篇,修正 %d 篇", scanned, fixed)
			return
		}

		//按主键游标分批扫描,只查计数字段
		var articles []models.ArticleModel
		err := global.DB.
			Select("id", "digg_count", "collect_count", "comment_count").
			Where("id > ?", cursor).
			Order("id asc").
			Limit(reconcileBatchSize).
			Find(&articles).Error
		if err != nil {
			logrus.Errorf("对账查询文章失败: %v", err)
			return
		}
		if len(articles) == 0 {
			logrus.Infof("文章计数对账完成,共扫描 %d 篇,修正 %d 篇", scanned, fixed)
			return
		}

		ids := make([]uint, 0, len(articles))
		for _, article := range articles {
			ids = append(ids, article.ID)
		}
		cursor = ids[len(ids)-1]
		scanned += len(ids)

		//关系表实际值
		diggActual := groupArticleCount(&models.ArticleDiggModel{}, ids)
		collectActual := groupArticleCount(&models.UserArticleCollectModel{}, ids)
		commentActual := groupArticleCount(&models.CommentModel{}, ids)
		//Redis中尚未同步的增量
		diggPending := redis_count.GetAllCacheDigg(ids)
		collectPending := redis_count.GetAllCacheCollect(ids)
		commentPending := redis_count.GetAllCacheComment(ids)

		fixedIDs := make([]uint, 0)
		for _, article := range articles {
			diggTarget := clampCount(diggActual[article.ID] - diggPending[article.ID])
			collectTarget := clampCount(collectActual[article.ID] - collectPending[article.ID])
			commentTarget := clampCount(commentActual[article.ID] - commentPending[article.ID])

			if article.DiggCount == diggTarget && article.CollectCount == collectTarget && article.CommentCount == commentTarget {
				continue
			}

			err = global.DB.Model(&models.ArticleModel{}).
				Where("id = ?", article.ID).
				Updates(map[string]any{
					"digg_count":    diggTarget,
					"collect_count": collectTarget,
					"comment_count": commentTarget,
				}).Error
			if err != nil {
				logrus.Errorf("修正文章 %d 计数失败: %v", article.ID, err)
				continue
			}
			logrus.Infof("修正文章 %d 计数: 点赞 %d->%d, 收藏 %d->%d, 评论 %d->%d",
				article.ID,
				article.DiggCount, diggTarget,
				article.CollectCount, collectTarget,
				article.CommentCount, commentTarget,
			)
			fixedIDs = append(fixedIDs, article.ID)
			fixed++
		}

		//修正后数据库计数已变化,失效详情缓存让前端读到新值
		redis_count.InvalidateArticleDetailCache(fixedIDs)

		time.Sleep(reconcileSleep)
	}
}

// groupArticleCount 按文章ID统计关系表记录数
func groupArticleCount(model any, ids []uint) map[uint]int {
	result := make(map[uint]int, len(ids))
	var rows []articleCountRow
	err := global.DB.Model(model).
		Select("article_id, count(*) as cnt").
		Where("article_id in ?", ids).
		Group("article_id").
		Scan(&rows).Error
	if err != nil {
		logrus.Errorf("统计文章关联记录失败: %v", err)
		return result
	}
	for _, row := range rows {
		result[row.ArticleID] = row.Cnt
	}
	return result
}

// clampCount 计数不允许出现负数
func clampCount(v int) int {
	if v < 0 {
		return 0
	}
	return v
}
