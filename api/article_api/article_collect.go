package article_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	"StarDreamerCyberNook/service/message_service"
	"StarDreamerCyberNook/service/redis_service/redis_count"
	jwts "StarDreamerCyberNook/utils/jwts"
	utils_other "StarDreamerCyberNook/utils/other"
	"StarDreamerCyberNook/utils/sql"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ArticleCollectRequest struct {
	ArticleID uint `json:"articleID" binding:"required"`
	CollectID uint `json:"collectID"`
}

func (ArticleApi) ArticleCollectView(c *gin.Context) {
	var req ArticleCollectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	var article models.ArticleModel
	err := global.DB.Take(&article, "status = ? and id = ?", models.StatusPublished, req.ArticleID).Error
	if err != nil {
		response.FailWithMsg("文章不存在", c)
		return
	}
	claims := jwts.GetClaims(c)

	// 确定目标收藏夹
	var collectModel models.CollectModel
	if req.CollectID == 0 {
		// 使用默认收藏夹,不存在时自动创建
		err = global.DB.Take(&collectModel, "user_id = ? and is_default = ?", claims.UserID, true).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			collectModel = models.CollectModel{Title: "默认收藏夹", UserID: claims.UserID, IsDefault: true}
			if err = global.DB.Create(&collectModel).Error; err != nil {
				response.FailWithMsg("创建默认收藏夹失败", c)
				return
			}
		} else if err != nil {
			response.FailWithMsg("查询默认收藏夹失败", c)
			return
		}
	} else {
		// 指定收藏夹必须存在且属于当前用户
		if err = global.DB.Take(&collectModel, "id = ? and user_id = ?", req.CollectID, claims.UserID).Error; err != nil {
			response.FailWithMsg("收藏夹不存在", c)
			return
		}
	}

	// 一个用户对一篇文章只保留一条收藏记录,收藏数统一由模型钩子维护,避免手动双写导致漂移
	var articleCollect models.UserArticleCollectModel
	err = global.DB.Take(&articleCollect, "user_id = ? and article_id = ?", claims.UserID, req.ArticleID).Error
	if err == nil {
		if articleCollect.CollectID == collectModel.ID {
			// 已收藏且是同一个收藏夹:取消收藏
			// 注:该中间表没有主键,必须显式带Where条件,同时把真实模型传给Delete让钩子拿到文章ID
			if err = global.DB.Where("user_id = ? and article_id = ?", claims.UserID, req.ArticleID).
				Delete(&articleCollect).Error; err != nil {
				response.FailWithMsg("取消收藏失败", c)
				return
			}
			response.OkWithMsg("取消收藏成功", c)
			return
		}
		// 已收藏但在别的收藏夹:移动到目标收藏夹,收藏总数不变
		// 注:该中间表没有主键,必须显式带Where条件
		if err = global.DB.Model(&models.UserArticleCollectModel{}).
			Where("user_id = ? and article_id = ?", claims.UserID, req.ArticleID).
			Update("collect_id", collectModel.ID).Error; err != nil {
			response.FailWithMsg("移动收藏夹失败", c)
			return
		}
		response.OkWithMsg("已移动到新的收藏夹", c)
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		response.FailWithMsg("查询收藏记录失败", c)
		return
	}

	// 收藏
	record := models.UserArticleCollectModel{
		UserID:    claims.UserID,
		ArticleID: req.ArticleID,
		CollectID: collectModel.ID,
	}
	if err = global.DB.Create(&record).Error; err != nil {
		response.FailWithMsg("收藏失败", c)
		return
	}
	response.OkWithMsg("收藏成功", c)
	// 发送收藏消息,失败只记录日志,不影响主流程
	if err = message_service.InsertCollectMessage(record); err != nil {
		logrus.Error("发送收藏消息失败", err.Error())
	}
}

type CollectCreateRequest struct { //创建收藏夹请求参数,请求创建时不用传id参数,除了创建也可以用于更新收藏夹
	Title    string `json:"title" binding:"required,max=32" s:"title"`
	Abstract string `json:"abstract" s:"abstract"`
	Cover    string `json:"cover" s:"cover"`
}

func (ArticleApi) CollectCreateView(c *gin.Context) {
	var req CollectCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	claims := jwts.GetClaims(c)
	//不允许创建重复名称的收藏夹
	// var model models.CollectModel
	// err := global.DB.Take(&model, "user_id = ? and title = ?", claims.UserID, req.Title).Error
	// if err == nil {
	// 	response.FailWithMsg("收藏夹名称重复", c)
	// 	return
	// }

	// 创建
	if global.DB.Create(&models.CollectModel{
		Title:    req.Title,
		UserID:   claims.UserID,
		Abstract: req.Abstract,
		Cover:    req.Cover,
	}).Error != nil {
		response.FailWithMsg("创建收藏夹失败", c)
		return
	}
	response.OkWithMsg("创建收藏夹成功", c)
}

type CollectUpdateRequest struct { //创建收藏夹请求参数,请求创建时不用传id参数,除了创建也可以用于更新收藏夹
	ID       uint    `json:"id" binding:"required"` //更新时需要传id参数
	Title    *string `json:"title" `
	Abstract *string `json:"abstract"`
	Cover    *string `json:"cover"`
}

func (ArticleApi) CollectUpdateView(c *gin.Context) {
	var req CollectUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	claims := jwts.GetClaims(c)
	var model models.CollectModel
	err := global.DB.Take(&model, "user_id = ? and id = ?", claims.UserID, req.ID).Error
	if err != nil {
		response.FailWithMsg("收藏夹不存在", c)
		return
	}

	updateMap := utils_other.StructToMap(&req, "sql") //把请求参数转换成map,方便后续更新,并且只更新有值的字段
	//也许不允许更新成重复名称的收藏夹?
	//现在来看还允许吧,给用户更高的自由度,毕竟收藏夹名称也不是很重要,而且用户也可以通过id和封面区分不同的收藏夹,所以就不做这个限制了

	err = global.DB.Model(&model).Updates(updateMap).Error
	if err != nil {
		response.FailWithMsg("更新收藏夹失败", c)
		return
	}

	response.OkWithMsg("更新收藏夹成功", c)
}

func (ArticleApi) CollectRemoveView(c *gin.Context) {
	var req models.RemoveRequest
	// 1. 参数绑定
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	if len(req.IDList) == 0 {
		response.OkWithMsg("未找到可删除的收藏夹或无权限", c)
		return
	}

	// 2. 基础查询构建 (注意：不要提前执行 Find)
	query := global.DB.Model(&models.CollectModel{}).
		Where("id IN ? AND is_default = ?", req.IDList, false)

	// 3. 权限控制
	claims := jwts.GetClaims(c)
	if claims.Role != enum.AdminRole {
		// 非管理员只能删除自己的收藏夹
		query = query.Where("user_id = ?", claims.UserID)
	}

	// 4. 先取出真正要删除的收藏夹ID,用于级联清理收藏记录
	var folderIDs []uint
	if err := query.Pluck("id", &folderIDs).Error; err != nil {
		response.FailWithMsg("查询收藏夹失败: "+err.Error(), c)
		return
	}
	if len(folderIDs) == 0 {
		// 情况 A: ID 列表为空
		// 情况 B: ID 不存在
		// 情况 C: 所有 ID 都是默认收藏夹 (is_default=true)
		// 情况 D: 非管理员尝试删除他人的收藏夹 (被 user_id 过滤)
		response.OkWithMsg("未找到可删除的收藏夹或无权限", c)
		return
	}

	// 5. 统计收藏夹内的收藏记录,按文章聚合后回退Redis计数
	var records []models.UserArticleCollectModel
	if err := global.DB.Where("collect_id IN ?", folderIDs).Find(&records).Error; err != nil {
		response.FailWithMsg("查询收藏记录失败", c)
		return
	}
	deltaMap := make(map[uint]int)
	for _, record := range records {
		deltaMap[record.ArticleID]--
	}

	// 6. 事务内删除收藏记录与收藏夹(跳过钩子,计数在上面统一回退)
	err := global.DB.Transaction(func(tx *gorm.DB) error {
		if len(records) > 0 {
			if err := tx.Session(&gorm.Session{SkipHooks: true}).
				Where("collect_id IN ?", folderIDs).
				Delete(&models.UserArticleCollectModel{}).Error; err != nil {
				return err
			}
		}
		return tx.Where("id IN ?", folderIDs).Delete(&models.CollectModel{}).Error
	})
	if err != nil {
		response.FailWithMsg("删除收藏夹失败: "+err.Error(), c)
		return
	}

	// 7. 数据库提交成功后再调整缓存计数
	for articleID, delta := range deltaMap {
		redis_count.SetCacheCollectBy(articleID, delta)
	}

	response.OkWithMsg("删除收藏夹成功", c)
}

type CollectListViewRequest struct {
	common.PageInfo
	ID uint `form:"id"`
}

// 先写着吧,一般来说是只有好友才能互看收藏夹的,或者以后给管理员看
func (ArticleApi) CollectListView(c *gin.Context) { //先看看用户有没有公开收藏夹,再查询用户收藏夹列表
	var req CollectListViewRequest //查询用户的收藏夹列表,应该使用分页查询吧

	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	if req.ID == 0 {
		response.FailWithMsg("参数错误", c)
		return
	}

	//检查收藏夹所属用户,查看用户是否开启收藏夹列表,如果没有开启且非用户,则返回错误,通过后分页查询收藏夹列表
	var user models.UserConfModel
	claims, _ := jwts.ParseTokenByGin(c)
	global.DB.Where("user_id = ?", req.ID).First(&user)                         //看看用户有没有开启收藏夹功能
	if user.OpenCollect != true && (claims == nil || claims.UserID != req.ID) { //如果用户没有开启收藏夹功能,并且请求者不是用户本人,则返回错误
		response.FailWithMsg("用户未开启收藏夹功能", c)
		return
	}

	//通过验证,组织数据,分页查询
	var query models.CollectModel

	var option common.Options
	//对req的option进行过滤
	option.Where = global.DB.Where("user_id = ?", req.ID) //只允许查询指定ID的收藏夹
	option.Likes = []string{"title", "abstract"}          //只允许模糊匹配收藏夹名称和摘要
	option.DefaultOrder = "created_at desc"
	option.AllowedOrders = []string{"id", "created_at"}

	data, count, err := common.ListQuery(query, option)
	if err != nil {
		response.FailWithMsg("查询收藏夹列表失败", c)
		return
	}
	// global.DB.Where("user_id = ?", req.ID).Find(&list)
	response.OkWithList(data, count, c)
}

type CollectArticleListViewRequest struct {
	common.PageInfo
	Likes []string `form:"likes"`
	ID    uint     `form:"id"`
}

func (ArticleApi) CollectArticleListView(c *gin.Context) {
	var req CollectArticleListViewRequest //查询用户的收藏夹文章列表
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	if req.ID == 0 {
		response.FailWithMsg("参数错误", c)
		return
	}

	//检查收藏夹所属用户,查看用户是否开启收藏夹列表
	var collect models.CollectModel
	if err := global.DB.Where("id = ?", req.ID).First(&collect).Error; err != nil {
		response.Fail(response.NotFound, "收藏夹不存在", response.EmptyData, c)
		return
	}
	claims, _ := jwts.ParseTokenByGin(c)
	//非本人(含未登录)访问他人收藏夹时,必须对方开启了公开收藏,否则统一按不存在处理,避免泄露隐私
	if claims == nil || claims.UserID != collect.UserID {
		var user models.UserConfModel
		global.DB.Where("user_id = ?", collect.UserID).First(&user)
		if user.OpenCollect != true {
			response.Fail(response.NotFound, "收藏夹不存在", response.EmptyData, c)
			return
		}
	}

	//通过验证,组织数据
	//收藏关系存放在用户收藏中间表因此先按收藏时间倒序取出文章ID,再查文章表
	var collectRecords []models.UserArticleCollectModel
	if err := global.DB.Where("collect_id = ?", req.ID).
		Order("created_at desc").Find(&collectRecords).Error; err != nil {
		response.FailWithMsg("查询收藏夹文章列表失败", c)
		return
	}
	if len(collectRecords) == 0 {
		response.OkWithList([]models.ArticleModel{}, 0, c)
		return
	}
	articleIDs := make([]uint, 0, len(collectRecords))
	for _, record := range collectRecords {
		articleIDs = append(articleIDs, record.ArticleID)
	}

	var option common.Options
	option.PageInfo = req.PageInfo                                                               //分页参数
	option.Where = global.DB.Where("id in ? and status = ?", articleIDs, models.StatusPublished) //只返回已发布文章
	option.Likes = []string{"title", "abstract"}                                                 //只允许模糊匹配文章标题和摘要
	option.DefaultOrder = sql.ConvertSliceOrderSql(articleIDs)                                   //保持收藏时间倒序
	option.AllowedOrders = []string{"id", "created_at", "look_count", "digg_count", "comment_count", "collect_count"}
	data, count, err := common.ListQuery(models.ArticleModel{}, option)
	if err != nil {
		response.FailWithMsg("查询收藏夹文章列表失败", c)
		return
	}

	//叠加Redis中未同步的计数增量,避免收藏/评论后要等定时任务回写才看到变化
	ptrs := make([]*models.ArticleModel, 0, len(data))
	for i := range data {
		ptrs = append(ptrs, &data[i])
	}
	applyArticleCountDeltas(ptrs)

	response.OkWithList(data, count, c)
}
