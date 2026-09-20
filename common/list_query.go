// common/list_query.go
package common

import (
	"StarDreamerCyberNook/global"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type PageInfo struct {
	EndId uint   `form:"endId"`              //某页末尾的游标ID,使用后可以减轻数据库压力
	Limit int    `form:"limit" default:"10"` //一页的数量,默认10条
	Page  int    `form:"page" default:"1"`   //分页查询,默认第1页
	Key   string `form:"key"`                //模糊匹配的参数
	Order string `form:"order"`              //其他查询参数,主要是排序
}

// GetOffset 获取查询偏移量
// 返回:int - 偏移量
// 说明:根据页码和每页数量计算偏移量
func (p *PageInfo) GetOffset() int {
	page := p.GetPage()
	limit := p.GetLimit()
	return (page - 1) * limit
}

// GetLimit 获取每页查询数量
// 返回:int - 每页数量限制
// 说明:限制范围在1到40之间,超出默认10
func (p *PageInfo) GetLimit() int {
	if p.Limit < 1 || p.Limit > 40 {
		return 10
	}
	return p.Limit
}

// GetPage 获取当前页码
// 返回:int - 当前页码
// 说明:页码范围在1到20之间,超出默认1
func (p *PageInfo) GetPage() int {
	if p.Page < 1 || p.Page > 20 {
		return 1
	}
	return p.Page
}

type Options struct { //可用选项,目前有模糊匹配和预加载,以后可以根据需要添加其他选项
	PageInfo      PageInfo //分页查询参数
	Likes         []string //模糊匹配的字段
	Preloads      []string //预加载的关联表
	Where         *gorm.DB //定制化查询
	DefaultOrder  string   //其他查询参数,主要是排序,前端没有传入就使用这个默认排序参数
	AllowedOrders []string //允许的排序字段白名单,前端传入的Order必须命中白名单,否则忽略
}

// TODO:AllowedOrders []string改为map[string]{}类型，提高效率

// sanitizeOrder 校验并规范化前端传入的排序参数,防止 ORDER BY 注入
// 参数:order - 前端传入的排序字符串,如 "created_at desc,id asc"
// 参数:allowed - 允许的字段白名单
// 返回:合法的排序字符串,不合法时返回空字符串
// 说明:只允许 "字段 [asc|desc]" 形式,且字段必须命中白名单,否则整体丢弃
func sanitizeOrder(order string, allowed []string) string {
	order = strings.TrimSpace(order)
	if order == "" || len(allowed) == 0 {
		return ""
	}
	allow := make(map[string]struct{}, len(allowed))
	for _, item := range allowed {
		allow[strings.ToLower(item)] = struct{}{}
	}
	parts := strings.Split(order, ",")
	safeParts := make([]string, 0, len(parts))
	for _, part := range parts {
		fields := strings.Fields(part)
		if len(fields) == 0 || len(fields) > 2 {
			return ""
		}
		column := strings.ToLower(fields[0])
		if _, ok := allow[column]; !ok {
			return ""
		}
		if len(fields) == 2 {
			dir := strings.ToLower(fields[1])
			if dir != "asc" && dir != "desc" {
				return ""
			}
			safeParts = append(safeParts, column+" "+dir)
		} else {
			safeParts = append(safeParts, column)
		}
	}
	return strings.Join(safeParts, ",")
}

// ListQuery 通用分页查询函数
// 参数:model - 模型实例,用于指定查询的表
// 参数:option - 查询的配置选项
// 返回:list - 查询结果列表
// 返回:count - 满足条件的总记录数
// 返回:err - 错误信息
// 说明:支持基础查询,模糊匹配,定制化查询,预加载,排序分页,游标分页
func ListQuery[T any](model T, option Options) (list []T, count int, err error) {
	// 基础查询
	baseQuery := global.DB.Model(model).Where(model)

	// 模糊匹配
	if len(option.Likes) > 0 && option.PageInfo.Key != "" {
		likeCond := global.DB.Session(&gorm.Session{NewDB: true})
		for i, column := range option.Likes {
			if i == 0 {
				likeCond = likeCond.Where(fmt.Sprintf("%s LIKE ?", column), "%"+option.PageInfo.Key+"%")
			} else {
				likeCond = likeCond.Or(fmt.Sprintf("%s LIKE ?", column), "%"+option.PageInfo.Key+"%")
			}
		}
		baseQuery = baseQuery.Where(likeCond)
	}

	// 定制化查询
	if option.Where != nil {
		baseQuery = baseQuery.Where(option.Where)
	}

	// 预加载
	for _, preLoad := range option.Preloads {
		baseQuery = baseQuery.Preload(preLoad)
	}

	// 构建最终查询（用于实际数据查询）
	finalQuery := baseQuery.Session(&gorm.Session{})

	// 排序（游标分页时需要唯一排序保证稳定性）
	// 前端传入的 Order 必须命中白名单,防止 ORDER BY 注入
	if order := sanitizeOrder(option.PageInfo.Order, option.AllowedOrders); order != "" {
		finalQuery = finalQuery.Order(order).Order("id DESC")
	} else if option.DefaultOrder != "" {
		finalQuery = finalQuery.Order(option.DefaultOrder).Order("id DESC")
	} else {
		finalQuery = finalQuery.Order("id DESC")
	}

	limit := option.PageInfo.GetLimit()

	// 游标分页模式（使用EndId）
	if option.PageInfo.EndId > 0 {
		finalQuery = finalQuery.Where("id < ?", option.PageInfo.EndId)

		// 查询总数（基于原始条件，不包含游标过滤）
		var total int64
		countQuery := baseQuery.Session(&gorm.Session{})
		if err = countQuery.Count(&total).Error; err != nil {
			return list, 0, err
		}
		count = int(total)

		err = finalQuery.Limit(limit).Find(&list).Error
		return list, count, err
	}

	//TODO:统计太耗时且无用了,考虑改进

	// 普通分页模式
	var total int64
	countQuery := baseQuery.Session(&gorm.Session{})
	if err = countQuery.Count(&total).Error; err != nil {
		return list, 0, err
	}
	count = int(total)

	offset := option.PageInfo.GetOffset()
	err = finalQuery.Offset(offset).Limit(limit).Find(&list).Error
	return list, count, err
}
