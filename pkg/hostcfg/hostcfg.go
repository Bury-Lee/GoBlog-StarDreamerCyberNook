// Package hostcfg 从全局配置(setting.yaml)派生各组件的默认参数,
// 使"一个 setting.yaml 说了算":蓝图只需声明启用哪些组件。
package hostcfg

import (
	"fmt"
	"strings"

	"StarDreamerCyberNook/global"
)

var currentSettingPath = "setting.yaml"

// SetSettingPath 记录当前使用的配置文件路径(供 blog 组件等派生)。
func SetSettingPath(p string) {
	if p != "" {
		currentSettingPath = p
	}
}

// SettingPath 返回当前配置文件路径。
func SettingPath() string { return currentSettingPath }

// OrStr 非空则取 v,否则取 def。
func OrStr(v, def string) string {
	if strings.TrimSpace(v) != "" {
		return v
	}
	return def
}

// OrInt 非零则取 v,否则取 def。
func OrInt(v, def int) int {
	if v != 0 {
		return v
	}
	return def
}

// OrFloat 非零则取 v,否则取 def。
func OrFloat(v, def float64) float64 {
	if v != 0 {
		return v
	}
	return def
}

// ServiceAddr 服务监听/地址(services.<name> 首个)。
func ServiceAddr(name, def string) string {
	if c := global.Config; c != nil {
		if a := c.Services[strings.ToLower(name)]; len(a) > 0 && strings.TrimSpace(a[0]) != "" {
			return a[0]
		}
	}
	return def
}

// DBDialect 从 dbWrite[0] 派生数据库驱动与 DSN。
func DBDialect() (driver, dsn string) {
	c := global.Config
	if c == nil || len(c.DBWrite) == 0 {
		return "sqlite", ""
	}
	d := c.DBWrite[0]
	switch strings.ToLower(string(d.SqlName)) {
	case "mysql":
		return "mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			d.User, d.Password, d.Host, d.Port, d.DBName)
	case "postgres", "postgresql":
		return "postgres", fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
			d.Host, d.User, d.Password, d.DBName, d.Port)
	default:
		return "sqlite", d.DBName
	}
}

// JWT 返回签发密钥与过期(网关/服务共用同一 JWT 配置)。
func JWT() (access, refresh string, accessMin, refreshHours int, issuer string) {
	if c := global.Config; c != nil {
		return c.Jwt.AccessTokenSecret, c.Jwt.RefreshTokenSecret, c.Jwt.AccessExpire, c.Jwt.RefreshExpire, c.Jwt.Issuer
	}
	return "", "", 0, 0, ""
}

// ES 返回 Elasticsearch 配置。
func ES() (enabled bool, url, user, pass string) {
	if c := global.Config; c != nil {
		return c.ES.Enabled, c.ES.Url, c.ES.UserName, c.ES.Password
	}
	return
}

// AI 返回 AI 配置。
func AI() (host, apiKey, apiType, model string, temperature float32, maxTokens int) {
	if c := global.Config; c != nil {
		return c.AI.Host, c.AI.ApiKey, c.AI.APIType, c.AI.Model, c.AI.Temperature, c.AI.MaxTokens
	}
	return "", "", "", "", 0, 0
}

// Email 返回邮件配置。
func Email() (domain string, port int, sendEmail, authCode, sendNickname string) {
	if c := global.Config; c != nil {
		return c.Email.Domain, c.Email.Port, c.Email.SendEmail, c.Email.AuthCode, c.Email.SendNickname
	}
	return "", 0, "", "", ""
}

// Media 返回媒体存储配置(对象存储启用时 backend=minio,否则 local)。
// 注意:图片 key 已包含 uploadDir 前缀(如 images/xxx),故本地根取 "."(运行根),
// 避免 hosts 目录与 key 前缀重复拼接。
func Media() (backend, localDir, host, accessKey, secretKey, bucket, region string) {
	c := global.Config
	if c == nil {
		return "local", ".", "", "", "", "", ""
	}
	backend, localDir = "local", "."
	if c.ObjectStorage.Enable {
		backend = "minio"
	}
	return backend, localDir, c.ObjectStorage.Host, c.ObjectStorage.AccessKey, c.ObjectStorage.SecretKey, c.ObjectStorage.Bucket, c.ObjectStorage.Region
}
