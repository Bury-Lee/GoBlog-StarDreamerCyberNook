package article_api

import (
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/redis_service/redis_count"
)

// applyArticleCountDeltas 把Redis中尚未同步到数据库的计数增量叠加到文章上
// 说明:浏览/点赞/收藏/评论计数都是先写Redis再由定时任务回写数据库,
// 若不叠加增量,用户操作后要等下一轮同步才能看到数字变化
func applyArticleCountDeltas(list []*models.ArticleModel) {
	if len(list) == 0 {
		return
	}
	ids := make([]uint, 0, len(list))
	for _, item := range list {
		if item == nil || item.ID == 0 {
			continue
		}
		ids = append(ids, item.ID)
	}
	if len(ids) == 0 {
		return
	}

	lookMap := redis_count.GetAllCacheLook(ids)
	diggMap := redis_count.GetAllCacheDigg(ids)
	collectMap := redis_count.GetAllCacheCollect(ids)
	commentMap := redis_count.GetAllCacheComment(ids)

	for _, item := range list {
		if item == nil || item.ID == 0 {
			continue
		}
		item.LookCount = clampCount(item.LookCount + lookMap[item.ID])
		item.DiggCount = clampCount(item.DiggCount + diggMap[item.ID])
		item.CollectCount = clampCount(item.CollectCount + collectMap[item.ID])
		item.CommentCount = clampCount(item.CommentCount + commentMap[item.ID])
	}
}

// applySearchCountDeltas 搜索结果(内嵌ArticleModel)的计数增量叠加
func applySearchCountDeltas(list []ArticleSearchListResponse) {
	if len(list) == 0 {
		return
	}
	ptrs := make([]*models.ArticleModel, 0, len(list))
	for i := range list {
		ptrs = append(ptrs, &list[i].ArticleModel)
	}
	applyArticleCountDeltas(ptrs)
}

// clampCount 计数不允许出现负数
func clampCount(v int) int {
	if v < 0 {
		return 0
	}
	return v
}
