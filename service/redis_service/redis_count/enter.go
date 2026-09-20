package redis_count

import "github.com/redis/go-redis/v9"

// 业界主流方案（建议版）

// - 方案 A（主流且落地快）： Redis 计数 + MQ/Stream 异步落库 + 对账
//   - 写入：点赞/收藏/浏览先做关系表/幂等校验，再原子 INCRBY 计数，同时投递事件到 Kafka/Redis Stream 。
//   - 落库：消费者按分区顺序消费，批量 UPDATE article SET count = count + ? ，失败重试+死信。
//   - 读取：详情静态信息单独缓存；计数独立读取 Redis（或短 TTL 本地缓存）。
//   - 保障：每日离线对账（关系表反算 vs article 计数），自动修正漂移。
// - 方案 B（高并发进阶）： 事件溯源 + 近实时聚合
//   - 所有行为事件只入日志流，不直接改 article 计数。
//   - 流处理（Flink/Kafka Streams）实时聚合写回 Redis/ClickHouse，MySQL 定时固化快照。
//   - 适合超大规模互动场景，复杂度更高。

type CacheType string

// articleCacheType Redis中文章统计缓存的哈希键类型
const (
	articleCacheLook    CacheType = "article_look_key"
	articleCacheDigg    CacheType = "article_digg_key"
	articleCacheCollect CacheType = "article_collect_key"
	articleCacheComment CacheType = "article_comment_key"

	DirtyArticleSetKey string = "cache_dirty_article_ids" //脏文章ID集合键

	commontCacheDigg CacheType = "comment_digg_key"

	DirtyCommentSetKey string = "cache_dirty_comment_ids" //脏评论ID集合键
)

// setDirtScript 原子地更新计数并标记脏数据,避免多命令之间存在"有增量无脏ID"的窗口
// KEYS[1]: 计数哈希键, KEYS[2]: 脏ID集合键, ARGV[1]: 字段(ID), ARGV[2]: 增量
var setDirtScript = redis.NewScript(`
redis.call('HINCRBY', KEYS[1], ARGV[1], ARGV[2])
redis.call('SADD', KEYS[2], ARGV[1])
return 1
`)

// ackScript 确认同步:按增量回退字段值,不会写入负值,归零后删除字段
// KEYS[1]: 计数哈希键, ARGV[1]: 字段(ID), ARGV[2]: 已同步的增量
var ackScript = redis.NewScript(`
local cur = tonumber(redis.call('HGET', KEYS[1], ARGV[1]) or '0')
local delta = tonumber(ARGV[2])
if delta < 0 then delta = -delta end
local next = cur - delta
if next < 0 then next = 0 end
if next == 0 then
  redis.call('HDEL', KEYS[1], ARGV[1])
else
  redis.call('HSET', KEYS[1], ARGV[1], next)
end
return next
`)
