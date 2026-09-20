package middleware

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// TODO:升级一下,改为对于不同的路由使用不同的限制,比如登录接口可以限制为每分钟5次,其他接口可以限制为每分钟100次,还可以根据用户ID进行限制,比如每个用户每天只能修改密码3次之类的

// ActLimitMiddleware 全局限流:同一IP每分钟最多64次请求
// 说明:使用独立前缀的key与原子INCR+EXPIRE,避免与其他限流共用key互相污染
func ActLimitMiddleware(c *gin.Context) {
	if !allowRequest(c, "act_limit:", 64, time.Minute) {
		response.FailWithMsg("请求过于频繁", c)
		c.Abort()
		return
	}
	c.Next()
}

// EmailSendLimitMiddleware 邮件发送限流:同一IP一分钟只能发一次
func EmailSendLimitMiddleware(c *gin.Context) {
	if !allowRequest(c, "email_send_limit:", 1, time.Minute) {
		response.FailWithMsg("请求过于频繁,请一分钟后再试", c)
		c.Abort()
		return
	}
	c.Next()
}

// ImgPostLimitMiddleware 图片上传限流:同一IP每分钟最多20次
func ImgPostLimitMiddleware(c *gin.Context) {
	if !allowRequest(c, "img_post_limit:", 20, time.Minute) {
		response.FailWithMsg("请求过于频繁", c)
		c.Abort()
		return
	}
	c.Next()
}

// allowRequest 基于Redis原子计数判断是否放行
// 参数:prefix - 限流维度key前缀,limit - 窗口内最大次数,window - 窗口时长
// 返回:true表示放行;Redis异常时放行(与既有行为一致)
func allowRequest(c *gin.Context, prefix string, limit int64, window time.Duration) bool {
	IP := c.ClientIP()
	ctx := context.Background()
	key := prefix + IP

	times, err := global.RedisTimeCache.Incr(ctx, key).Result()
	if err != nil {
		logrus.Errorf("限流计数失败, key: %s, err: %v", key, err)
		return true
	}
	if times == 1 {
		if err := global.RedisTimeCache.Expire(ctx, key, window).Err(); err != nil {
			logrus.Errorf("限流过期时间设置失败, key: %s, err: %v", key, err)
		}
	}
	return times <= limit
}
