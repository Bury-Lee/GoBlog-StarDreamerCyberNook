package sql

import (
	"StarDreamerCyberNook/global"
	"fmt"
	"strconv"
	"strings"
)

// ConvertSliceOrderSql 按全局配置的数据库类型,生成"保持给定 ID 顺序"的排序片段(网关进程用)。
// 注意:独立服务进程没有 global.Config,应改用 ConvertSliceOrderSqlBy 并传入自身 driver。
func ConvertSliceOrderSql(orderList []uint) string {
	dbType := ""
	if global.Config != nil && len(global.Config.DBWrite) > 0 {
		dbType = string(global.Config.DBWrite[0].SqlName)
	}
	return ConvertSliceOrderSqlBy(dbType, orderList)
}

// ConvertSliceOrderSqlBy 按给定数据库类型生成排序片段(不依赖 global 配置)。
func ConvertSliceOrderSqlBy(dbType string, orderList []uint) string {
	if len(orderList) == 0 {
		return ""
	}

	switch strings.ToLower(dbType) {
	case "mysql", "mariadb":
		// MySQL: FIELD(id, 1, 2, 3)
		idStrs := make([]string, 0, len(orderList))
		for _, id := range orderList {
			idStrs = append(idStrs, strconv.FormatUint(uint64(id), 10))
		}
		return fmt.Sprintf("FIELD(id, %s)", strings.Join(idStrs, ","))

	default:
		// PostgreSQL / SQLite / 通用: CASE id WHEN 1 THEN 0 ... ELSE n END
		var sb strings.Builder
		sb.WriteString("CASE id ")
		for i, id := range orderList {
			sb.WriteString(fmt.Sprintf("WHEN %d THEN %d ", id, i))
		}
		sb.WriteString(fmt.Sprintf("ELSE %d END", len(orderList)))
		return sb.String()
	}
}
