package middleware

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/utils/jwts"
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type EmailVerifyInfoRequest struct {
	EmailID   string `json:"emailID" binding:"required"`
	EmailCode string `json:"emailCode" binding:"required"`
}

type EmailVerifyInfo struct {
	RequstEmail string `json:"requestEmail"` //请求的邮箱,预备字段,不从请求体获取,而是从邮箱存储里获取,验证成功后存入上下文
	EmailID     string `json:"emailID" binding:"required"`
	EmailCode   string `json:"emailCode" binding:"required"`
	Type        string `json:"type" binding:"required"` //注册,重置密码,重置邮箱
}

// EmailVerifyMiddleware 邮箱验证中间件
// 参数:c - gin上下文对象
// 说明:验证邮箱验证码,获取请求体并解析,调用邮箱存储验证,验证通过后将邮箱信息存入上下文
func EmailVerifyMiddleware(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		response.FailWithMsg("获取请求体错误", c)
		c.Abort()
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	var req EmailVerifyInfoRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		logrus.Errorf("邮箱验证失败 %s", err)
		response.FailWithMsg("邮箱验证失败", c)
		c.Abort()
		return
	}

	info, ok := verifyEmailCode(c, fmt.Sprintf("email:%s", req.EmailID), req.EmailCode, "", "")
	if !ok {
		c.Abort()
		return
	}

	c.Set("email", info)

	c.Request.Body = io.NopCloser(bytes.NewReader(body))
}

// emailCodeMaxAttempts 验证码最大尝试次数,超过后作废验证码
const emailCodeMaxAttempts = 5

// verifyEmailCode 读取并校验邮箱验证码,校验通过后原子消费,防止重放与暴力破解
// 参数:key - Redis中的验证码键,code - 用户提交的验证码
// 参数:expectEmail - 非空时要求验证码所属邮箱一致
// 参数:expectType - 非空时要求验证码业务类型一致
// 返回:验证码信息与是否通过
func verifyEmailCode(c *gin.Context, key, code, expectEmail, expectType string) (EmailVerifyInfo, bool) {
	ctx := context.Background()
	data, err := global.RedisTimeCache.Get(ctx, key).Result()
	if err != nil {
		response.FailWithMsg("数据异常", c)
		return EmailVerifyInfo{}, false
	}

	var info EmailVerifyInfo
	if err := json.Unmarshal([]byte(data), &info); err != nil {
		response.FailWithMsg("数据异常", c)
		return EmailVerifyInfo{}, false
	}

	if expectEmail != "" && info.RequstEmail != expectEmail {
		response.FailWithMsg("邮箱验证出错", c)
		return EmailVerifyInfo{}, false
	}
	if expectType != "" && info.Type != expectType {
		response.FailWithMsg("验证出错", c)
		return EmailVerifyInfo{}, false
	}

	if subtle.ConstantTimeCompare([]byte(info.EmailCode), []byte(code)) != 1 {
		//失败计数,超过上限直接作废验证码
		failKey := key + ":fail"
		attempts, incrErr := global.RedisTimeCache.Incr(ctx, failKey).Result()
		if incrErr == nil && attempts == 1 {
			global.RedisTimeCache.Expire(ctx, failKey, 10*time.Minute)
		}
		if attempts >= emailCodeMaxAttempts {
			global.RedisTimeCache.Del(ctx, key)
		}
		response.FailWithMsg("验证码错误", c)
		return EmailVerifyInfo{}, false
	}

	//原子消费验证码,防止并发重复使用
	consumed, err := global.RedisTimeCache.GetDel(ctx, key).Result()
	if err != nil || consumed != data {
		response.FailWithMsg("验证码已失效,请重新获取", c)
		return EmailVerifyInfo{}, false
	}
	global.RedisTimeCache.Del(ctx, key+":fail")
	return info, true
}

type ResetEmailVerifyInfoRequest struct { //用于验证原邮箱是否通过的请求体结构体
	EmailID   string `json:"ResetEmailID" binding:"required"`
	EmailCode string `json:"ResetEmailCode" binding:"required"`
}

// ResetEmailVerifyMiddleware 邮箱验证中间件
// 参数:c - gin上下文对象
// 说明:验证邮箱验证码,获取请求体并解析,调用邮箱存储验证,验证通过后将邮箱信息存入上下文
func ResetEmailVerifyMiddleware(c *gin.Context) { //专门给重置邮箱用的中间件,要验证两次,一次验证原邮箱,一次验证新邮箱
	//这里仅检验原邮箱是否通过,不检验新邮箱是否通过
	//只关心请求者是否有原邮箱的验证码,不关心新邮箱的验证码是否正确
	body, err := c.GetRawData()
	if err != nil {
		response.FailWithMsg("获取请求体错误", c)
		c.Abort()
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	var req ResetEmailVerifyInfoRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		logrus.Errorf("邮箱验证失败 %s", err)
		response.FailWithMsg("邮箱验证失败", c)
		c.Abort()
		return
	}

	//校验原邮箱是否属于当前登录用户,防止验证码被盗用
	user, err := jwts.GetClaims(c).GetUser()
	if err != nil {
		response.FailWithMsg("获取用户信息失败", c)
		c.Abort()
		return
	}
	if _, ok := verifyEmailCode(c, fmt.Sprintf("ResetEmail:%s", req.EmailID), req.EmailCode, user.Email, "重置邮箱"); !ok {
		c.Abort()
		return
	}

	c.Request.Body = io.NopCloser(bytes.NewReader(body))
}
