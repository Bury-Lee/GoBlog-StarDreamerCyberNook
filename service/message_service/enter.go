package message_service

import (
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"errors"
)

// InsertCommentMessage 给文章作者发送评论消息(评论创建成功后调用,消息里带上评论ID)
func InsertCommentMessage(model models.CommentModel, RevUserID uint) error {
	var actionUser models.UserModel
	if err := global.DB.Select("id", "nick_name", "avatar").Take(&actionUser, model.UserID).Error; err != nil {
		return err
	}
	var article models.ArticleModel
	if err := global.DB.Select("id", "title").Take(&article, model.ArticleID).Error; err != nil {
		return err
	}
	err := global.DB.Create(&models.MessageModel{
		Type:               models.MessageTypeComment,
		RevUserID:          RevUserID,
		ActionUserID:       model.UserID,
		ActionUserNickname: actionUser.NickName,
		ActionUserAvatar:   actionUser.Avatar,
		Title:              models.MessageTypeComment.String(),
		ArticleID:          model.ArticleID,
		ArticleTitle:       article.Title,
		CommentID:          model.ID,
		Content:            model.Content,
		IsRead:             false,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

// InsertReplyMessage 给被回复的评论作者发送回复消息(评论创建成功后调用,消息里带上评论ID)
func InsertReplyMessage(model models.CommentModel, RevUserID uint) error {
	var actionUser models.UserModel
	if err := global.DB.Select("id", "nick_name", "avatar").Take(&actionUser, model.UserID).Error; err != nil {
		return err
	}
	var article models.ArticleModel
	if err := global.DB.Select("id", "title").Take(&article, model.ArticleID).Error; err != nil {
		return err
	}
	err := global.DB.Create(&models.MessageModel{
		Type:               models.MessageTypeReply,
		RevUserID:          RevUserID,
		ActionUserID:       model.UserID,
		ActionUserNickname: actionUser.NickName,
		ActionUserAvatar:   actionUser.Avatar,
		Title:              models.MessageTypeReply.String(),
		ArticleID:          model.ArticleID,
		ArticleTitle:       article.Title,
		CommentID:          model.ID,
		IsRead:             false,
		Content:            model.Content,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

func InsertArticleDiggMessage(model models.ArticleDiggModel) error { //给文章的作者发送点赞消息
	global.DB.Preload("ArticleModel").Take(&model)
	//要查询点赞的人的用户信息
	var ActionUser models.UserModel
	global.DB.Where("id = ?", model.UserID).Take(&ActionUser)
	err := global.DB.Create(&models.MessageModel{
		Type:               models.MessageTypeDigg,
		RevUserID:          model.ArticleModel.UserID,
		ActionUserID:       ActionUser.ID,
		ActionUserNickname: ActionUser.NickName,
		ActionUserAvatar:   ActionUser.Avatar,
		Title:              models.MessageTypeDigg.String(),
		ArticleID:          model.ArticleID,
		ArticleTitle:       model.ArticleModel.Title,
		IsRead:             false,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

// TODO:以后加入给评论的作者发送点赞消息
func InsertCommentDiggMessage(model models.CommentModel, RevUserID uint) error { //给评论的人发送点赞消息
	return errors.New("TODO")
}

func InsertCollectMessage(model models.UserArticleCollectModel) error { //给文章的人发送收藏消息
	global.DB.Preload("ArticleModel").Preload("UserModel").Take(&model)
	//要查询点赞的人的用户信息
	err := global.DB.Create(&models.MessageModel{
		Type:               models.MessageTypeCollect,
		RevUserID:          model.ArticleModel.UserID,
		ActionUserID:       model.UserID,
		ActionUserNickname: model.UserModel.NickName,
		ActionUserAvatar:   model.UserModel.Avatar,
		Title:              models.MessageTypeCollect.String(),
		ArticleID:          model.ArticleID,
		ArticleTitle:       model.ArticleModel.Title,
		IsRead:             false,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

//TODO:以后加入给关注的人发送关注消息

func InsertSystemMessage(message models.MessageModel) error { //给别人发送系统消息,例如审核通过,账号被冻结等
	err := global.DB.Create(&models.MessageModel{
		Type:               models.MessageTypeSystem,
		RevUserID:          message.RevUserID,
		ActionUserID:       message.ActionUserID,
		ActionUserNickname: message.ActionUserNickname,
		ActionUserAvatar:   message.ActionUserAvatar,
		Title:              message.Title,
		ArticleID:          message.ArticleID,
		ArticleTitle:       message.ArticleTitle,
		CommentID:          message.CommentID,
		Content:            message.Content,
		LinkTitle:          message.LinkTitle,
		LinkHref:           message.LinkHref,
		IsRead:             false,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

func InsertAtMessage(model models.UserModel, ReceverUserID uint) error { //给别人发送@消息
	err := global.DB.Create(&models.MessageModel{
		Type:               models.MessageTypeAt,
		RevUserID:          ReceverUserID,
		ActionUserID:       model.ID,
		ActionUserNickname: model.NickName,
		ActionUserAvatar:   model.Avatar,
		Title:              models.MessageTypeAt.String(),
		IsRead:             false,
	}).Error
	if err != nil {
		return err
	}
	return nil
}
