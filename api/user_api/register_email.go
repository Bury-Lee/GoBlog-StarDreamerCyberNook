package user_api

//TODO:如果没记错的话邮箱验证码只能验证一次,输错一次就直接作废,到时候改一下,改为10次
import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/middleware"
	"StarDreamerCyberNook/models/enum"
	"StarDreamerCyberNook/service/ai_service"
	"StarDreamerCyberNook/service/user_service"
	"StarDreamerCyberNook/utils"
	xss_filter "StarDreamerCyberNook/utils/XSSfilter"
	Hash "StarDreamerCyberNook/utils/hash"
	jwts "StarDreamerCyberNook/utils/jwts"
	utils_other "StarDreamerCyberNook/utils/other"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RegisterEmailRequest 邮箱注册请求参数结构体
// 用于接收前端提交的注册信息，包含邮箱验证ID、验证码、密码和昵称
type RegisterEmailRequest struct {
	EmailID   string `json:"emailID" binding:"required"`   // 邮箱验证记录的唯一标识ID
	EmailCode string `json:"emailCode" binding:"required"` // 邮箱收到的验证码
	Pwd       string `json:"password" binding:"required"`  // 用户设置的登录密码
	NickName  string `json:"nickName"`                     // 用户昵称（可选）
}

// RegisterEmailView 邮箱注册接口
// 处理用户通过邮箱验证码完成的注册流程
// 流程：验证参数 -> 校验邮箱验证码 -> 生成用户名 -> 创建用户 -> 返回Token
func (UserApi) RegisterEmailView(c *gin.Context) {
	var req RegisterEmailRequest
	// 1. 参数绑定与基础校验
	// 将JSON请求体绑定到结构体，并校验必填字段
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数绑定失败", c)
		return
	}
	//昵称为纯文本,先做xss清洗
	req.NickName = xss_filter.SanitizeText(req.NickName)

	//如果启用ai审核
	if global.Config.AI.Enable { //启用ai审核
		res, err := ai_service.CreateSingleReply(req.NickName, global.SystemPromptUser.String())
		if err != nil {
			logrus.Errorf("ai审核失败: %s", err.Error())
			//出错自动降级为非ai流程
		}
		switch res {
		case "通过":
			//通过,更新用户信息
		case "拒绝":
			//拒绝,返回错误
			response.FailWithMsg("违规昵称,请重试", c)
			return
		default:
			//注意:禁止打印整个请求体,避免明文密码进入日志
			logrus.Errorf("ai审核结果未知: %s,昵称:%s", res, req.NickName)
			response.FailWithMsg("违规昵称,请重试", c) //也许也可以考虑放行?
			return
		}
	}

	// 2. 查找邮箱验证记录
	// 从Redis中根据EmailID获取之前发送验证码时存储的记录
	data, exit := c.Get("email")
	if !exit {
		response.FailWithMsg("获取不到邮箱验证信息", c)
		return
	}
	info, ok := data.(middleware.EmailVerifyInfo)
	if !ok {
		response.FailWithMsg("邮箱验证信息格式错误", c)
		return
	}
	if info.Type != "注册" {
		response.FailWithMsg("邮箱验证类型错误", c)
		return
	}

	// 6. 生成随机用户名
	// 使用32位数字字符串作为系统用户名（非登录用，仅内部标识）
	uname := utils.GetRandomString(32, utils.Digits)

	// 7. 密码哈希加密
	// 使用bcrypt等算法对密码进行单向哈希，确保数据库不存储明文密码
	hashPwd, _ := Hash.HashPassword(req.Pwd)

	// 8. 邮箱格式标准化
	// 将邮箱转换为小写，确保邮箱唯一性判断不受大小写影响
	info.RequstEmail = utils_other.ToLower(info.RequstEmail)

	// 9. 建号(下沉 user 服务;服务内做邮箱唯一性二次校验,防验证码重放/并发重复注册)
	userID, userName, role, err := user_service.CreateUserByEmail(
		info.RequstEmail, "用户"+req.NickName, uname, hashPwd, enum.UserRole)
	if err != nil {
		if errors.Is(err, user_service.ErrEmailUsed) {
			response.FailWithMsg("该邮箱已注册", c)
			return
		}
		logrus.Errorf("用户创建失败: %s", err) // 记录详细错误日志供排查
		response.FailWithMsg("注册失败,可能是邮箱已注册或其他错误", c)
		return
	}

	// 11. 检查邮箱登录功能是否开启
	// 如果后台关闭了邮箱登录，提示用户记住生成的用户名（用于账号密码登录）
	if !global.Config.Site.Login.EmailLogin {
		response.FailWithMsg(fmt.Sprintf("目前暂时不支持邮箱登录,您的用户名是:%s,请您牢记", userName), c)
	}

	// 12. 生成JWT令牌
	// 注册成功后自动登录，生成包含用户ID、用户名、角色的Token
	AccessToken, RefreshToken, err := jwts.GetToken(jwts.Claims{
		UserID:   userID,
		Username: userName,
		Role:     role,
	})
	if err != nil {
		response.FailWithMsg("登录失败", c)
		return
	}

	// 13. 返回成功响应
	// 将JWT令牌返回给前端，后续请求携带此Token进行身份验证
	response.OkWithData(map[string]string{
		"AccessToken":  AccessToken,
		"RefreshToken": RefreshToken,
	}, c)
}
