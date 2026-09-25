package flags

import (
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"os"

	"github.com/sirupsen/logrus"
)

func FlagDB() { //数据库迁移
	err := global.DB.AutoMigrate(
		&models.UserModel{},
		&models.UserConfModel{},
		&models.ArticleDiggModel{},
		&models.ArticleModel{},
		&models.CategoryModel{},
		&models.CollectModel{},
		&models.UserArticleCollectModel{},
		&models.UserArticleHistoryModel{},
		&models.ImageModel{},
		&models.CommentModel{},
		&models.LogModel{},
		&models.BannerModel{},
		&models.GlobalNotificationModel{},
		&models.FriendLink{},
		&models.FriendPromotion{},
		&models.UserLoginModel{},
		&models.UserMessageConfModel{},
		&models.MessageModel{},
		&models.UserFollowModel{},
		&models.ChatModel{},
		&models.TextMsg{},
		&models.ImageMsg{},
		&models.MarkdownMsg{},
		&models.ChatMsg{},
		&models.UserChatActionModel{},
		&models.ArticleSearchModel{},
		&models.UserTopArticleModel{},
		&models.CommentDiggModel{},
		&models.SessionModel{},
		&models.ArticleSearchModel{},
		&models.ArticleAddition{},
		&models.FeedbackModel{},
		&models.MomentModel{},
		&models.MomentDiggModel{},
		&models.MomentCommentModel{},
		&models.MomentCommentDiggModel{},
	)
	if err != nil {
		//迁移失败必须以非0退出码结束,避免脚本/CI误判成功
		logrus.Errorf("数据库迁移失败 %s", err)
		os.Exit(1)
	}
	logrus.Info("数据库已迁移")
}
