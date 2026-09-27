// utils/jwts/access_token.go
package jwts

import (
	"errors"
	"fmt"

	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// Claims 自定义JWT声明结构体
type Claims struct {
	UserID   uint          `json:"ID"`   // 用户唯一标识符
	Username string        `json:"name"` // 用户名
	Role     enum.RoleType `json:"role"` // 用户角色类型
}

// MyClaims 继承自定义声明并嵌入标准声明
type MyClaims struct {
	Claims
	jwt.RegisteredClaims
}

// GetUser 根据JWT中的UserID从数据库中获取用户完整信息
func (this *MyClaims) GetUser() (models.UserModel, error) {
	var user models.UserModel
	err := global.DB.Take(&user, this.UserID).Error
	return user, err
}

var jwtSigningMethod = jwt.SigningMethodHS256

// GetAccessToken 签发访问令牌(经 auth 服务)。
func GetAccessToken(claims Claims) (string, error) {
	c, err := client()
	if err != nil {
		return "", err
	}
	ctx, cancel := ctxTimeout()
	defer cancel()
	rep, err := c.IssueAccess(ctx, issueReq(claims.UserID, claims.Username, int(claims.Role)))
	if err != nil {
		return "", err
	}
	return rep.GetAccessToken(), nil
}

// ParseAccessToken 解析和验证JWT令牌(本地校验,热路径不远程)。
func ParseAccessToken(tokenString string) (*MyClaims, error) {
	if tokenString == "" {
		return nil, errors.New("请登录")
	}

	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwtSigningMethod {
			logrus.Errorf("算法混淆攻击: %v!", token.Header["alg"])
			return nil, fmt.Errorf("非法的算法: %v", token.Header["alg"])
		}
		return []byte(global.Config.Jwt.AccessTokenSecret), nil
	})
	if err != nil {
		logrus.Errorf("token解析失败: %v", err)
		return nil, err
	}

	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

// ParseTokenByGin 从Gin上下文解析JWT(优先请求头 token,其次 URL 查询参数)。
func ParseTokenByGin(c *gin.Context) (*MyClaims, error) {
	token := c.GetHeader("token")
	if token == "" {
		token = c.Query("token")
	}
	return ParseAccessToken(token)
}

// GetClaims 从Gin上下文获取已解析的JWT声明。
func GetClaims(c *gin.Context) *MyClaims {
	_claims, ok := c.Get("claims")
	if !ok {
		return nil
	}
	claims, ok := _claims.(*MyClaims)
	if !ok {
		return nil
	}
	return claims
}
