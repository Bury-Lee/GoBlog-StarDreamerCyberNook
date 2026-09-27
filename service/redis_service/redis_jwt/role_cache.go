// service/redis_service/redis_jwt/role_cache.go
package redis_jwt

import (
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models/enum"
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// 用户角色/封禁状态缓存:鉴权中间件每个受保护请求都要知道最新角色,
// 缓存后避免每请求查库。封禁/改角色时通过 InvalidateRoleCache 主动失效,保证即时生效;
// TTL 仅作为遗漏失效时的兜底。
const (
	roleCachePrefix = "user_role_"
	roleCacheTTL    = 10 * time.Minute
)

func roleCacheKey(userID uint) string {
	return fmt.Sprintf("%s%d", roleCachePrefix, userID)
}

// GetRoleCache 读取缓存的用户角色;ok=false 表示未命中或 Redis 不可用。
func GetRoleCache(userID uint) (enum.RoleType, bool) {
	if global.RedisTimeCache == nil {
		return 0, false
	}
	v, err := global.RedisTimeCache.Get(context.Background(), roleCacheKey(userID)).Int()
	if err != nil {
		return 0, false
	}
	return enum.RoleType(v), true
}

// SetRoleCache 写入用户角色缓存(含封禁角色,便于快速拒绝)。
func SetRoleCache(userID uint, role enum.RoleType) {
	if global.RedisTimeCache == nil {
		return
	}
	if err := global.RedisTimeCache.Set(context.Background(), roleCacheKey(userID), int(role), roleCacheTTL).Err(); err != nil {
		logrus.Errorf("写入用户角色缓存失败 userID=%d: %v", userID, err)
	}
}

// InvalidateRoleCache 清除用户角色缓存,使封禁/角色变更即时生效。
func InvalidateRoleCache(userID uint) {
	if global.RedisTimeCache == nil {
		return
	}
	if err := global.RedisTimeCache.Del(context.Background(), roleCacheKey(userID)).Err(); err != nil {
		logrus.Errorf("清除用户角色缓存失败 userID=%d: %v", userID, err)
	}
}
