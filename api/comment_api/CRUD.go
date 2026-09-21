package comment_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	"StarDreamerCyberNook/service/ai_service"
	"StarDreamerCyberNook/service/message_service"
	"StarDreamerCyberNook/service/redis_service/redis_count"
	xss_filter "StarDreamerCyberNook/utils/XSSfilter"
	jwts "StarDreamerCyberNook/utils/jwts"
	utils_other "StarDreamerCyberNook/utils/other"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// 注:评论没有修改这一说,只有增删查
type CommentCreateRequest struct {
	Content   string `json:"content" binding:"required"`
	ArticleID uint   `json:"articleID" binding:"required"`
	ParentID  uint   `json:"parentID"` // 父评论ID
}

// 对于更新就由前端来做吧,收到成功响应之后直接更新
// TODO:创建和删除评论文章的评论数统计在redis进行,需要在创建和删除评论时更新缓存
func (CommentApi) CommentCreateView(c *gin.Context) {
	//流程:参数和文章合规性检验
	//填充内容
	//设置评论关系+发消息
	var req CommentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	var article models.ArticleModel                                                                       //必须是发布的文章才能发评论
	err := global.DB.Take(&article, "id = ? and status = ?", req.ArticleID, models.StatusPublished).Error //TODO:调试时用,发布后要去掉Debug()
	if err != nil {
		response.FailWithMsg("文章不存在", c)
		return
	}
	claims := jwts.GetClaims(c)

	//评论内容防xss注入
	req.Content = xss_filter.NewXSSFilter().Sanitize(req.Content)
	if req.Content == "" {
		response.FailWithMsg("评论内容不能为空", c)
		return
	}

	model := models.CommentModel{ //当前评论模型
		Content:   req.Content,
		UserID:    claims.UserID,
		ArticleID: req.ArticleID,
	}

	if global.Config.AI.Enable { //启用ai审核
		reply, err := ai_service.CreateSingleReply(
			"评论内容:"+req.Content,
			global.SystemPromptComment.String(),
		)
		if err != nil {
			response.FailWithMsg("ai审核失败,已经自动创建为待审核状态: "+err.Error(), c)
			return
		}
		switch reply { //TODO:这里无论成功还是失败都应该插入消息,告知原因
		case "通过":
			//通过就正常执行流程
		case "拒绝":
			response.FailWithMsg("评论可能含有违规信息,已拒绝", c)
			return
		default:
			logrus.Errorf("ai审核出错,回复内容:%s,评论详情:%s\n,评论不发布", reply, fmt.Sprintf("%#v", req))
			return
		}
	}

	//收集需要发送回复消息的接收人,等评论创建成功后再统一发送(消息里要带上评论ID)
	var replyRevUserIDs []uint
	if req.ParentID == 0 {
		// 父评论路径为空,说明这个评论是一级评论
		model.RootParentID = nil
		model.ParentPath = ""
	}
	// 否则这是二级评论(或更深层级的回复)
	if req.ParentID != 0 {
		//不对,现在无论是二级评论怎么样,都应该是:model.RootParentID=Parentmodel.Parentpath的/.../,而父评论是最后一个路径,这样哪怕是次级回复也可以有正确的逻辑
		//为了节省空间,入库时应该以base64编码计入
		var parentModel models.CommentModel
		err := global.DB.Take(&parentModel, "id = ? and article_id = ?", req.ParentID, req.ArticleID).Error //TODO:调试时用,发布后要去掉Debug()
		if err != nil {
			response.FailWithMsg("评论不存在", c) // 回复的评论得是这个文章里的
			return
		}
		model.ParentPath = utils_other.EncodePath(parentModel.ParentPath, parentModel.ID)
		//给父评论发消息(自己回复自己不发送)
		if parentModel.UserID != claims.UserID {
			replyRevUserIDs = append(replyRevUserIDs, parentModel.UserID)
		}
		if parentModel.RootParentID == nil {
			// 如果父评论本身是根评论 (RootParentID 为 nil)，则当前评论的根即为父评论,发布的新评论为二级评论
			model.RootParentID = &parentModel.ID
		} else {
			// 如果父评论不是根评论，则直接继承其根评论ID，无需重新解码,发布的新评论为二级的次级评论
			model.RootParentID = parentModel.RootParentID

			//这里只能再查一次数据库看看根评论的用户ID来获取根评论的用户ID
			var rootParentModel models.CommentModel
			err = global.DB.Take(&rootParentModel, "id = ? and article_id = ?", parentModel.RootParentID, req.ArticleID).Error //走到这一步,说明这是二级评论,二级评论的根评论不可能不存在
			if err != nil {
				logrus.Errorf("系统消息发送失败,无法查询.父评论ID: %v, 文章ID: %v", parentModel.RootParentID, req.ArticleID)
			} else if rootParentModel.UserID != parentModel.UserID && rootParentModel.UserID != claims.UserID { //如果根评论的用户ID和父评论的用户ID不一样,说明是不同的人,就给根评论发消息,如果是同一个人就不发了,避免重复发消息了
				replyRevUserIDs = append(replyRevUserIDs, rootParentModel.UserID)
			}
		}
	}

	//先落库,再发消息:消息需要用到评论ID,创建失败时也不会产生幽灵消息
	err = global.DB.Create(&model).Error
	if err != nil {
		response.FailWithMsg("发布评论失败", c)
		return
	}

	for _, revUserID := range replyRevUserIDs {
		if err = message_service.InsertReplyMessage(model, revUserID); err != nil {
			logrus.Error("系统消息发送失败:", err.Error())
		}
	}
	//给文章作者发消息(评论自己的文章不发送)
	if article.UserID != claims.UserID {
		if err = message_service.InsertCommentMessage(model, article.UserID); err != nil {
			logrus.Error("系统消息发送失败:", err.Error())
		}
	}

	//旧方法,直接打到数据库,现在改为先更新缓存,然后定时任务再批量更新到数据库
	// if global.DB.Model(&models.ArticleModel{}).Where("id = ?", req.ArticleID).Select("comment_count").Updates(map[string]interface{}{
	// 	"comment_count": gorm.Expr("comment_count + ?", 1),
	// }).Error != nil {
	// 	logrus.Error("文章评论数更新失败,文章ID:", req.ArticleID)
	// }
	// 发布评论后,需要更新文章的评论数

	redis_count.SetCacheComment(req.ArticleID, true) //增量更新缓存

	response.OkWithMsg("发布评论成功", c)
}

// 分页获取指定文章的一级评论
type CommentDetailRequest struct {
	common.PageInfo
	ArticleID uint `form:"articleID" binding:"required"` // 文章ID
}

/*
方案：只缓存“评论 ID 列表” (推荐)
不要把完整的评论内容存入 ZSET，ZSET 只存 排序索引。
Redis ZSET 结构：
Key: article:comments:hot:{article_id}
Member: comment_id (例如: 1001, 1005)
Score: digg_count (点赞数) 或者 timestamp (如果是按时间排序)
注意：如果是按点赞数排序，Score 会动态变化，需要频繁更新 ZSET 的 Score。
缓存策略：
第一步： 请求进来，先去 Redis 执行 ZREVRANGE key 0 19 (获取前20个热门评论ID)。
第二步： 拿到 ID 列表后，去 Redis 的 String/Hash 缓存中批量获取评论详情（HMGET comments:detail {id1} {id2}...）。
第三步： 如果详情缓存未命中，再去数据库查，查完后回填缓存。
优点：
ZSET 非常轻量，只存 ID 和分数。
更新点赞数时，只需 ZINCRBY，不需要移动整个对象。
*/
func (CommentApi) CommentListlView(c *gin.Context) { //获取某文章的一级评论(分页)
	//TODO:redis缓存
	var req CommentDetailRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	//先检查文章状态
	var article models.ArticleModel
	err := global.DB.Take(&article, "id = ? and status = ?", req.ArticleID, models.StatusPublished).Error //TODO:调试时用,发布后要去掉Debug()
	if err != nil {
		response.FailWithMsg("文章不存在", c)
		return
	}

	var comments models.CommentModel
	var options common.Options
	options.PageInfo = req.PageInfo
	options.Preloads = []string{"UserModel"}                                                    //预加载用户信息
	options.Where = global.DB.Where("article_id = ? and root_parent_id is null", req.ArticleID) //查询一级评论
	options.DefaultOrder = "digg_count desc"                                                    //默认按点赞数降序排序
	options.AllowedOrders = []string{"id", "created_at", "digg_count"}
	List, count, err := common.ListQuery(comments, options)
	if err != nil { //TODO:加一个点赞增量也加上的逻辑
		response.FailWithMsg("查询评论失败", c)
		return
	}
	for i, v := range List {
		var UserModel = models.UserModel{
			Model:         v.UserModel.Model,
			NickName:      v.UserModel.NickName,
			Avatar:        v.UserModel.Avatar,
			LastLoginTime: v.UserModel.LastLoginTime,
			Age:           v.UserModel.Age,
			LikeTags:      v.UserModel.LikeTags,
		}

		List[i].UserModel = UserModel
	}
	//叠加Redis里的点赞增量,否则点赞后要等10分钟定时任务回写才看得到变化
	if len(List) > 0 {
		ids := make([]uint, 0, len(List))
		for _, v := range List {
			ids = append(ids, v.ID)
		}
		diggMap := redis_count.GetAllCacheCommentDigg(ids)
		for i := range List {
			List[i].DiggCount += diggMap[List[i].ID]
		}
	}
	response.OkWithList(List, count, c)
}

// 分页获取某条评论下的子评论详情(多条)
type CommentListRequest struct {
	common.PageInfo
	Root uint `form:"root" binding:"required"` //根评论ID
}

func (CommentApi) CommentChildListView(c *gin.Context) { //可以这样,评论被删除了内容就替换为评论已删除
	var req CommentListRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	var comments models.CommentModel
	//先检查父评论是否存在
	err := global.DB.Take(&comments, "id = ?", req.Root).Error
	if err != nil {
		response.FailWithMsg("评论不存在", c)
		return
	}

	//检查文章状态
	var article models.ArticleModel
	err = global.DB.Take(&article, "id = ? and status = ?", comments.ArticleID, models.StatusPublished).Error
	if err != nil {
		response.FailWithMsg("文章不存在", c)
		return
	}

	var options common.Options
	options.PageInfo = req.PageInfo
	options.Preloads = []string{"UserModel"} //预加载用户信息
	//路径精确匹配:等值命中直接子评论,或"路径/"前缀命中更深层评论,避免 LIKE '/1%' 串到 '/11' 的子树
	path := utils_other.EncodePath(comments.ParentPath, comments.ID)
	options.Where = global.DB.Where(
		"article_id = ? and (parent_path = ? or parent_path like ?)",
		comments.ArticleID,
		path,
		path+"/%",
	) //直接使用父评论的文章ID,查询有相同根评论且父路径包含当前评论路径的所有评论
	//考虑两个问题,
	//1.如果父评论被删除了,那么子评论的parent_path就会失效,但是子评论的root_parent_id还会指向根评论,这时候就需要根据root_parent_id来查询
	//如果用户传入的Rootid是乱写的怎么办,放回err吧
	//用户的
	options.DefaultOrder = "digg_count desc" //默认按点赞数降序排序
	options.AllowedOrders = []string{"id", "created_at", "digg_count"}
	List, count, err := common.ListQuery(models.CommentModel{}, options)
	if err != nil {
		response.FailWithMsg("查询评论失败", c)
		return
	}
	for i, v := range List {
		var UserModel = models.UserModel{
			Model:         v.UserModel.Model,
			NickName:      v.UserModel.NickName,
			Avatar:        v.UserModel.Avatar,
			LastLoginTime: v.UserModel.LastLoginTime,
			Age:           v.UserModel.Age,
			LikeTags:      v.UserModel.LikeTags,
		}

		List[i].UserModel = UserModel
	}
	//叠加Redis里的点赞增量,否则点赞后要等10分钟定时任务回写才看得到变化
	if len(List) > 0 {
		ids := make([]uint, 0, len(List))
		for _, v := range List {
			ids = append(ids, v.ID)
		}
		diggMap := redis_count.GetAllCacheCommentDigg(ids)
		for i := range List {
			List[i].DiggCount += diggMap[List[i].ID]
		}
	}
	response.OkWithList(List, count, c)
}

// 删除指定评论及其子评论(?其实子评论不一定要删?
// TODO:创建和删除评论文章的评论数统计在redis进行,需要在创建和删除评论时更新缓存
func (CommentApi) CommentDeleteView(c *gin.Context) {
	var req models.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	var comment models.CommentModel
	err := global.DB.Take(&comment, "id = ?", req.ID).Error
	if err != nil {
		response.FailWithMsg("评论不存在", c)
		return
	}
	//先检查评论是否存在
	//检查文章状态
	var article models.ArticleModel
	err = global.DB.Take(&article, "id = ? and status = ?", comment.ArticleID, models.StatusPublished).Error //TODO:调试时用,发布后要去掉Debug()
	if err != nil {
		response.FailWithMsg("文章不存在", c)
		return
	}
	//鉴权,检查用户是否有删除评论的权限,只能删自己的,管理员可以删除别人的
	claim := jwts.GetClaims(c)
	if claim.UserID != comment.UserID && claim.Role != enum.AdminRole {
		response.FailWithMsg("没有权限删除评论", c)
		return
	}

	if comment.RootParentID == nil {
		//一级评论,连带着二级评论一起删除
		var childCount int64
		if err := global.DB.Model(&models.CommentModel{}).Where("root_parent_id = ?", comment.ID).Count(&childCount).Error; err != nil {
			response.FailWithMsg("查询子评论失败", c)
			return
		}
		//先删子评论,再删自己
		if err := global.DB.Delete(&models.CommentModel{}, "root_parent_id = ?", comment.ID).Error; err != nil {
			response.FailWithMsg("删除子评论失败", c)
			return
		}
		if err := global.DB.Delete(&comment).Error; err != nil {
			response.FailWithMsg("删除评论失败", c)
			return
		}
		//评论数只调整Redis增量,由定时任务回写数据库,避免数据库与增量被重复扣减
		redis_count.SetCacheCommentBy(comment.ArticleID, -int(childCount+1))
	} else { //只删除自己
		if err := global.DB.Delete(&comment).Error; err != nil {
			response.FailWithMsg("删除评论失败", c)
			return
		}
		redis_count.SetCacheCommentBy(comment.ArticleID, -1)
	}

	response.OkWithMsg("删除评论成功", c)
}

// 评论点赞
func (CommentApi) CommentDiggView(c *gin.Context) {
	var req models.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	var comment models.CommentModel
	err := global.DB.Take(&comment, req.ID).Error
	if err != nil {
		response.FailWithMsg("评论不存在", c)
		return
	}

	claim := jwts.GetClaims(c) //要先登录才能点赞
	digged := false
	if global.DB.Take(&models.CommentDiggModel{}, "user_id = ? and comment_id = ?", claim.UserID, req.ID).Error == gorm.ErrRecordNotFound {
		//查询不到说明没有点赞过,可以点赞
		if global.DB.Create(&models.CommentDiggModel{
			UserID:    claim.UserID,
			CommentID: req.ID,
		}).Error != nil {
			response.FailWithMsg("点赞失败", c)
			return
		}
		redis_count.SetCacheCommentDigg(req.ID, true) //增量加一
		digged = true
	} else {
		// 查询到说明已经点赞过,取消点赞
		if global.DB.Delete(&models.CommentDiggModel{}, "user_id = ? and comment_id = ?", claim.UserID, req.ID).Error != nil {
			response.FailWithMsg("取消点赞失败", c)
			return
		}
		redis_count.SetCacheCommentDigg(req.ID, false) //增量减一
	}

	//返回最新点赞状态与点赞数(数据库值+Redis增量),前端可直接更新界面
	diggCount := comment.DiggCount
	if delta, ok := redis_count.GetAllCacheCommentDigg([]uint{req.ID})[req.ID]; ok {
		diggCount += delta
	}
	if diggCount < 0 {
		diggCount = 0
	}
	response.OkWithData(gin.H{"digged": digged, "diggCount": diggCount}, c)
}
