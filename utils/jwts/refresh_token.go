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

// RefreshClaims 刷新令牌声明(仅携带用户ID与标准声明)。
type RefreshClaims struct {
	ID uint `json:"id"`
	jwt.RegisteredClaims
}

// GetRefreshToken 签发刷新令牌(经 auth 服务)。
func GetRefreshToken(userID uint) (string, error) {
	c, err := client()
	if err != nil {
		return "", err
	}
	ctx, cancel := ctxTimeout()
	defer cancel()
	rep, err := c.IssueRefresh(ctx, issueRefreshReq(userID))
	if err != nil {
		return "", err
	}
	return rep.GetRefreshToken(), nil
}

// ParseRefreshToken 解析刷新令牌(本地校验)。
func ParseRefreshToken(tokenString string) (*RefreshClaims, error) {
	if tokenString == "" {
		return nil, errors.New("请登录")
	}

	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwtSigningMethod {
			logrus.Errorf("非法的算法: %v", token.Header["alg"])
			return nil, fmt.Errorf("非法的算法: %v", token.Header["alg"])
		}
		return []byte(global.Config.Jwt.RefreshTokenSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*RefreshClaims)
	if !ok || !token.Valid {
		return nil, errors.New("无效的token")
	}
	return claims, nil
}

// GetToken 签发访问+刷新令牌(经 auth 服务)。
func GetToken(claims Claims) (string, string, error) {
	c, err := client()
	if err != nil {
		return "", "", err
	}
	ctx, cancel := ctxTimeout()
	defer cancel()
	rep, err := c.Issue(ctx, issueReq(claims.UserID, claims.Username, int(claims.Role)))
	if err != nil {
		return "", "", err
	}
	return rep.GetAccessToken(), rep.GetRefreshToken(), nil
}

// RefreshAccessToken 刷新通行令牌:本地校验 refresh + 查库校验用户状态,再经 auth 服务签发新 access。
func RefreshAccessToken(refreshToken string) (string, error) {
	rc, err := ParseRefreshToken(refreshToken)
	if err != nil {
		return "", err
	}

	var user models.UserModel
	if err = global.DB.Take(&user, rc.ID).Error; err != nil {
		return "", errors.New("用户不存在或已失效")
	}
	if user.Role == enum.BlackRole {
		return "", errors.New("用户已被封禁")
	}

	return GetAccessToken(Claims{
		UserID:   user.ID,
		Username: user.UserName,
		Role:     user.Role,
	})
}

// ParseRefreshTokenByGin 仅从请求体获取刷新令牌。
func ParseRefreshTokenByGin(c *gin.Context) (*RefreshClaims, error) {
	token := c.PostForm("token")
	if token == "" {
		return nil, errors.New("请登录")
	}
	return ParseRefreshToken(token)
}

// GetRefreshClaims 从Gin上下文获取刷新声明。
func GetRefreshClaims(c *gin.Context) *RefreshClaims {
	_claims, ok := c.Get("claims")
	if !ok {
		return nil
	}
	claims, ok := _claims.(*RefreshClaims)
	if !ok {
		return nil
	}
	return claims
}
