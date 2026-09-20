package core

import (
	"StarDreamerCyberNook/conf"
	"StarDreamerCyberNook/flags"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func ReadConf() *conf.Config {
	byteData, err := os.ReadFile(flags.FlagOptions.File) //读取文件
	if err != nil {
		panic(err) //以后试试改为"无法读取配置文件"+err.Error()
	}
	var c = new(conf.Config)
	err = yaml.Unmarshal(byteData, c) //结构绑定
	if err != nil {
		panic(fmt.Sprintf("yaml文件格式错误 %s", err))
	}
	//设置版本号
	c.Site.About.SetVersion()
	validateJWTSecret(c)
	return c
}

// validateJWTSecret 校验JWT密钥强度
// 说明:Debug模式下仅警告,Release模式下直接Panic,防止使用占位符/弱密钥上线
func validateJWTSecret(c *conf.Config) {
	check := func(name, secret string) {
		if !isWeakJWTSecret(secret) {
			return
		}
		if c.System.RunMode == "release" {
			logrus.Panicf("%s 不安全(为空/过短/占位符),请在配置中更换为至少32位的随机密钥", name)
		}
		logrus.Warnf("%s 不安全(为空/过短/占位符),生产环境必须更换为至少32位的随机密钥", name)
	}
	check("accessTokenSecret", c.Jwt.AccessTokenSecret)
	check("refreshTokenSecret", c.Jwt.RefreshTokenSecret)
}

func isWeakJWTSecret(secret string) bool {
	if len(secret) < 32 {
		return true
	}
	lower := strings.ToLower(secret)
	for _, placeholder := range []string{"xxxx", "123456", "password", "secret", "changeme"} {
		if strings.Contains(lower, placeholder) {
			return true
		}
	}
	return false
}
