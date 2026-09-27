package user_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/service/ai_service"
	"StarDreamerCyberNook/service/user_service"
	xss_filter "StarDreamerCyberNook/utils/XSSfilter"
	jwts "StarDreamerCyberNook/utils/jwts"
	utils_other "StarDreamerCyberNook/utils/other"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type UserInfoUpdateRequest struct {
	Avatar      *string            `json:"avatar" s-u:"avatar"`
	Abstract    *string            `json:"abstract" s-u:"abstract"`
	LikeTags    *[]string          `json:"likeTags" s-u:"like_tags"`
	NickName    *string            `json:"nickName" s-u:"nick_name"`       // 昵称
	Age         *int               `json:"Age" s-u:"age"`                  // 年龄
	ContactInfo *map[string]string `json:"contactInfo" s-u:"contact_info"` // 联系方式，JSON格式存储
	// Email       *string            `json:"email" s-u:"email"`              // 邮箱，唯一索引//这个和登录账号相关,暂时不放在这里更新

	OpenFollow  *bool `json:"openFollow" s-u-c:"open_follow"`    // 公开我的关注
	OpenFans    *bool `json:"openFans" s-u-c:"open_fans"`        // 公开我的粉丝
	HomeStyleID *uint `json:"homeStyleID" s-u-c:"home_style_id"` // 主页样式的id
}

func (UserApi) UserInfoUpdateView(c *gin.Context) {
	var req UserInfoUpdateRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMsg(err.Error(), c)
		return
	}

	if req.LikeTags != nil && len(*req.LikeTags) > 36 {
		response.FailWithMsg("喜欢的标签数量太多啦", c)
		return
	}
	//昵称/简介为纯文本,先做xss清洗
	if req.NickName != nil {
		nickname := xss_filter.SanitizeText(*req.NickName)
		req.NickName = &nickname
	}
	if req.Abstract != nil {
		abstract := xss_filter.SanitizeText(*req.Abstract)
		req.Abstract = &abstract
	}
	//ai审核环节
	if global.Config.AI.Enable { //启用ai审核
		content := ""
		if req.Abstract != nil {
			content += "\n用户简介:" + *req.Abstract
		}
		if req.NickName != nil {
			content += "\n用户昵称:" + *req.NickName
		}
		if req.LikeTags != nil {
			content += "\n用户喜欢的标签:" + fmt.Sprintf("%v", *req.LikeTags)
		}
		if req.ContactInfo != nil {
			content += "\n用户联系方式:" + fmt.Sprintf("%v", *req.ContactInfo)
		}
		//只有本次提交包含需要审核的文本内容时才调用AI,避免只改隐私开关也被AI审核拦截
		if content != "" {
			res, err := ai_service.CreateSingleReply(content, global.SystemPromptUser.String())
			if err != nil {
				logrus.Errorf("ai审核失败: %s", err.Error())
				//出错自动降级为非ai流程
			}
			switch res { //TODO:这里无论成功还是失败都应该插入消息,告知原因
			case "通过":
				//通过,更新用户信息
			case "拒绝":
				//拒绝,返回错误
				response.FailWithMsg("存在违规信息,用户信息未更新", c)
				return
			default:
				logrus.Errorf("ai审核结果未知: %s,用户:%+v", res, req)
				response.FailWithMsg("审核结果未知,用户信息未更新", c) //也许也可以考虑放行?
				return
			}
		}
	}

	userMap := utils_other.StructToMap(req, "s-u")
	userConfMap := utils_other.StructToMap(req, "s-u-c")

	var userPatch, confPatch string
	if len(userMap) > 0 {
		b, err := json.Marshal(userMap)
		if err != nil {
			response.FailWithMsg("用户信息修改失败", c)
			return
		}
		userPatch = string(b)
	}
	if len(userConfMap) > 0 {
		b, err := json.Marshal(userConfMap)
		if err != nil {
			response.FailWithMsg("用户信息修改失败", c)
			return
		}
		confPatch = string(b)
	}

	claims := jwts.GetClaims(c)
	if userPatch != "" || confPatch != "" {
		if err := user_service.UpdateUser(claims.UserID, userPatch, confPatch); err != nil {
			response.FailWithMsg("用户信息修改失败", c)
			return
		}
	}
	response.OkWithMsg("用户信息修改成功", c)
}
