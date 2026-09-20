package flags

import (
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"os"

	"github.com/sirupsen/logrus"
)

// FlagSearch 重建数据库降级搜索表
// 说明:先清空 article_search_models,再把全部已发布文章重新写入;
// 用于历史数据回填,或搜索表与文章表状态不一致时的修复
func FlagSearch() {
	if err := global.DB.Where("1 = 1").Delete(&models.ArticleSearchModel{}).Error; err != nil {
		logrus.Errorf("清空搜索表失败 %s", err)
		os.Exit(1)
	}

	var published int64
	if err := global.DB.Model(&models.ArticleModel{}).Where("status = ?", models.StatusPublished).Count(&published).Error; err != nil {
		logrus.Errorf("统计已发布文章失败 %s", err)
		os.Exit(1)
	}

	const batchSize = 500
	var lastID uint
	var inserted int64
	for {
		var list []models.ArticleModel
		if err := global.DB.Where("status = ? and id > ?", models.StatusPublished, lastID).
			Order("id asc").Limit(batchSize).Find(&list).Error; err != nil {
			logrus.Errorf("查询文章失败 %s", err)
			os.Exit(1)
		}
		if len(list) == 0 {
			break
		}
		records := make([]models.ArticleSearchModel, 0, len(list))
		for _, article := range list {
			records = append(records, models.NewArticleSearchModel(article))
			lastID = article.ID
		}
		if err := global.DB.CreateInBatches(records, batchSize).Error; err != nil {
			logrus.Errorf("写入搜索表失败 %s", err)
			os.Exit(1)
		}
		inserted += int64(len(records))
	}

	logrus.Infof("搜索表重建完成, 已发布文章 %d 篇, 写入搜索记录 %d 条", published, inserted)
}
