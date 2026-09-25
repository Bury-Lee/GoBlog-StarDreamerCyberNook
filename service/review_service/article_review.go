// service/review_service/article_review.go
// 文章审核共用逻辑:人工审核、接口AI审核与定时任务AI审核共用,保证各链路的缓存清理与通知行为一致
package review_service

import (
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/ai_service"
	"StarDreamerCyberNook/service/message_service"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	// maxConsecutiveAIFailures 连续AI调用失败达到该次数时提前终止本次审核,避免AI服务不可用时反复无效请求
	maxConsecutiveAIFailures = 10
	// autoReviewDeadline 定时任务单次执行的最长时间,避免待审核文章过多时长时间占用资源
	autoReviewDeadline = 30 * time.Minute
)

// ArticleAIReviewItem AI审核结果
type ArticleAIReviewItem struct {
	ArticleID uint          `json:"articleID"`
	Title     string        `json:"title"`
	AIResult  string        `json:"aiResult"` //AI的原始判定:通过/拒绝/其它
	Status    models.Status `json:"status"`   //审核后的文章状态
	Error     string        `json:"error,omitempty"`
}

// ApplyArticleReview 更新文章审核状态、清理详情缓存并给作者发送系统通知
// 说明:人工审核、接口AI审核与定时AI审核共用
func ApplyArticleReview(article *models.ArticleModel, status models.Status, msg string) error {
	if err := global.DB.Model(article).Update("status", status).Error; err != nil {
		return err
	}
	article.Status = status

	//启用AI时,审核通过(变为已发布)顺带刷新文章的AI点评,写入扩展附录表(记录AI模型名)
	//说明:AI定时审核(SyncAIReview)、接口AI审核与人工审核共用此处,保证点评与审核结果一致
	if status == models.StatusPublished && global.Config.AI.Enable {
		if quality, summary, err := ai_service.CommentArticle(article.Title, article.Abstract, article.Content); err != nil {
			logrus.Errorf("更新文章 %d 的AI点评失败: %v", article.ID, err)
		} else if err := global.DB.Where(models.ArticleAddition{ArticleID: article.ID}).
			Assign(map[string]any{
				"ai_quality":  quality,
				"ai_abstract": summary,
				"ai_model":    global.Config.AI.Model,
			}).
			FirstOrCreate(&models.ArticleAddition{}).Error; err != nil {
			logrus.Errorf("保存文章 %d 的AI点评失败: %v", article.ID, err)
		}
	}

	//状态变更后清理详情缓存,避免已下线文章继续可见或新通过的文章一直不可见
	global.RedisHotPool.Del(context.Background(), "ArticleID"+strconv.FormatUint(uint64(article.ID), 10))

	var content string
	switch status {
	case models.StatusPublished:
		content = fmt.Sprintf("文章通过审核:%s\n备注:%s", article.Title, msg)
	case models.StatusDraft:
		content = fmt.Sprintf("文章不通过审核:%s\n备注:%s", article.Title, msg)
	case models.StatusOffline:
		content = fmt.Sprintf("文章已下线:%s\n备注:%s", article.Title, msg)
	case models.StatusPending:
		content = fmt.Sprintf("文章进入待审核状态:%s\n备注:%s", article.Title, msg)
	}

	message := models.MessageModel{
		RevUserID:          article.UserID,
		ActionUserID:       0,
		ActionUserNickname: "系统",
		ActionUserAvatar:   "", //这里可以改用站内默认头像
		Title:              "文章审核通知",
		ArticleID:          article.ID,
		ArticleTitle:       article.Title,
		Content:            content,
	}
	if err := message_service.InsertSystemMessage(message); err != nil {
		logrus.Error("发送系统消息失败", err)
	}
	return nil
}

// reviewOneArticle 对单篇文章执行AI审核并按判定结果落库
// 返回:审核结果项,包含AI原始判定、落库后的状态与错误信息
func reviewOneArticle(article *models.ArticleModel) ArticleAIReviewItem {
	item := ArticleAIReviewItem{ArticleID: article.ID, Title: article.Title, Status: article.Status}

	reply, err := ai_service.ReviewArticle(article.Title, article.Abstract, article.Content)
	if err != nil {
		item.Error = err.Error()
		return item
	}
	item.AIResult = reply
	switch reply {
	case ai_service.AIReviewPass:
		item.Status = models.StatusPublished
	case ai_service.AIReviewReject:
		item.Status = models.StatusDraft
	default:
		//AI无法判定,保持待审核状态,交给人工处理
		item.Status = models.StatusPending
	}

	//状态没有变化时无需写库与发送通知,避免定时任务每轮都推送"进入待审核状态"
	if item.Status == article.Status {
		return item
	}

	if err = ApplyArticleReview(article, item.Status, "AI审核结果:"+reply); err != nil {
		item.Error = err.Error()
	}
	return item
}

// HasPendingArticles 是否存在待审核文章,供定时任务在探测AI服务前先行判断,避免无意义的探活请求
func HasPendingArticles() (bool, error) {
	var count int64
	if err := global.DB.Model(&models.ArticleModel{}).Where("status = ?", models.StatusPending).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ReviewPendingArticles 对处于"审核中"状态的文章执行AI审核
// 参数:idList - 指定文章ID,为空时审核全部待审核文章
// 参数:limit - 本次最多审核的篇数
// 审核规则:通过->已发布, 拒绝->草稿(退回作者), 无法判定->保持审核中等待人工处理
func ReviewPendingArticles(idList []uint, limit int) ([]ArticleAIReviewItem, error) {
	query := global.DB.Where("status = ?", models.StatusPending)
	if len(idList) > 0 {
		query = query.Where("id in ?", idList)
	}
	var articles []models.ArticleModel
	if err := query.Order("id asc").Limit(limit).Find(&articles).Error; err != nil {
		return nil, err
	}

	list := make([]ArticleAIReviewItem, 0, len(articles))
	for i := range articles {
		item := reviewOneArticle(&articles[i])
		if item.Error != "" {
			logrus.Errorf("AI审核文章 %d 失败: %s", item.ArticleID, item.Error)
		}
		list = append(list, item)
	}
	return list, nil
}

// ReviewAllPendingArticles 定时任务用:按ID游标分批审核全部待审核文章
// 参数:batchSize - 每批处理的文章数
// 参数:sleep - 每批之间的休眠,降低AI服务与数据库压力
// 返回:total 本次处理的文章数, success 状态更新成功数, failed 处理失败数, err 查询失败
// 说明:AI服务中途不可用(连续失败达上限)或超过最长执行时间时提前结束,剩余文章留待下次任务
func ReviewAllPendingArticles(batchSize int, sleep time.Duration) (total, success, failed int, err error) {
	if batchSize <= 0 {
		batchSize = 5
	}
	deadline := time.Now().Add(autoReviewDeadline)
	var cursor uint //游标:上一批处理到的最大ID,保证每篇文章单次任务只处理一次
	consecutiveFailures := 0

	for {
		if time.Now().After(deadline) {
			logrus.Warnf("自动AI审核超时,提前结束,本次已处理 %d 篇", total)
			return total, success, failed, nil
		}

		var articles []models.ArticleModel
		err = global.DB.
			Where("status = ? and id > ?", models.StatusPending, cursor).
			Order("id asc").
			Limit(batchSize).
			Find(&articles).Error
		if err != nil {
			return total, success, failed, err
		}
		if len(articles) == 0 {
			return total, success, failed, nil
		}

		for i := range articles {
			item := reviewOneArticle(&articles[i])
			cursor = item.ArticleID
			total++
			if item.Error == "" {
				success++
				consecutiveFailures = 0
			} else {
				failed++
				consecutiveFailures++
				logrus.Errorf("自动AI审核文章 %d 失败: %s", item.ArticleID, item.Error)
				time.Sleep(4 * time.Second) //4秒钟等待
				if consecutiveFailures >= maxConsecutiveAIFailures {
					logrus.Warnf("AI服务连续 %d 次不可用,提前结束本次自动审核,剩余文章待下次任务处理", consecutiveFailures)
					return total, success, failed, nil
				}
			}
		}

		//每批之间休眠,避免AI服务与数据库压力过大
		time.Sleep(sleep)
	}
}
