package es_service

import (
	"context"
	"time"

	searchv1 "StarDreamerCyberNook/gen/search/v1"
)

// Hit 是搜索结果条目(标题/摘要可能已高亮)。
type Hit struct {
	ID       uint64
	Title    string
	Abstract string
}

// SearchArticles 经 search 服务检索文章。
// index 索引名;key 关键词;tag 标签;sortType 排序(0..4);
// offset/size 分页;status 状态过滤;topIDs 置顶加权;likeTags 兴趣标签(type=1)。
func SearchArticles(index, key, tag string, sortType int8, offset, size, status int, topIDs []uint, likeTags []string) (int64, []Hit, error) {
	c, err := client()
	if err != nil {
		return 0, nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	top := make([]uint64, 0, len(topIDs))
	for _, id := range topIDs {
		top = append(top, uint64(id))
	}

	rep, err := c.SearchArticles(ctx, &searchv1.SearchArticlesRequest{
		Index:    index,
		Key:      key,
		Tag:      tag,
		Type:     int32(sortType),
		Offset:   int32(offset),
		Size:     int32(size),
		Status:   int32(status),
		TopIds:   top,
		LikeTags: likeTags,
	})
	if err != nil {
		return 0, nil, err
	}

	hits := make([]Hit, 0, len(rep.GetHits()))
	for _, h := range rep.GetHits() {
		hits = append(hits, Hit{ID: h.GetId(), Title: h.GetTitle(), Abstract: h.GetAbstract()})
	}
	return rep.GetTotal(), hits, nil
}
