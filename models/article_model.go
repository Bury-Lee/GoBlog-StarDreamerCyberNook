// models/article_model.go
// 文章模型定义
// 存储博客系统的文章信息，包括标题、内容、分类、标签、统计等完整字段
package models

import (
	"database/sql/driver"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type TagList []string

func (this *TagList) Scan(value any) error {
	val, ok := value.([]uint8)
	if ok {
		*this = strings.Split(string(val), ",")
		return nil
	}
	return errors.New(fmt.Sprint("断言失败:", value))
}

func (this TagList) Value() (driver.Value, error) {
	return strings.Join(this, ","), nil
}

// ArticleModel 文章模型
// 用于存储博客系统的文章信息，包含标题、内容、分类、标签、统计等字段
type ArticleModel struct {
	Model
	Title         string         `gorm:"size:32" json:"title"`                     // 文章标题，最大32字符
	Abstract      string         `gorm:"size:256" json:"abstract"`                 // 文章摘要，最大256字符
	Content       string         `json:"content"`                                  // 文章内容
	CategoryID    *uint          `json:"categoryID"`                               // 文章分类ID，关联分类表
	CategoryModel *CategoryModel `gorm:"foreignKey:CategoryID" json:"-"`           //这样可以吗? 以防万一用指针吧             // 分类信息，外键关联，不序列化
	TagList       []string       `gorm:"type:text;serializer:json" json:"tagList"` // 标签列表，JSON序列化存储 //serializer:json要删掉?似乎要换成自己定义的taglist数据类型
	Cover         string         `gorm:"size:256" json:"cover"`                    // 文章封面图片URL
	UserID        uint           `json:"userID"`                                   // 作者用户ID，关联用户表
	UserModel     UserModel      `gorm:"foreignKey:UserID" json:"-"`               // 作者信息，外键关联，不序列化
	LookCount     int            `json:"lookCount"`                                // 浏览次数统计
	DiggCount     int            `json:"diggCount"`                                // 点赞次数统计
	CommentCount  int            `json:"commentCount"`                             // 评论数量统计
	CollectCount  int            `json:"collectCount"`                             // 收藏次数统计//TODO:每个用户对于每篇文章只能收藏一次,不然搞多几个收藏夹就会一直刷收藏统计,所以这个收藏数直接用redis缓存就好了?不需要每次都更新数据库了,有并发问题
	OpenComment   bool           `json:"openComment"`                              // 是否开启评论：true-开启 false-关闭
	Status        Status         `json:"status"`                                   // 文章状态：0-草稿 1-审核中 2-已发布 3-已下线

	//文章扩展附录,管理员附加评论,AI点评等独立成表,通过外键关联
	ArticleAddition *ArticleAddition `gorm:"foreignKey:ArticleID" json:"articleAddition"`
}
type Status int8 // 文章状态枚举类型

const (
	StatusDraft     Status = 0 // 草稿//注意的是默认就是零,当前端为传入指定参数时
	StatusPending   Status = 1 // 审核中 (Pending 比 Auditing 更常用于表示“等待处理”的状态)
	StatusPublished Status = 2 // 已发布
	StatusOffline   Status = 3 // 已下线 (或者用 StatusArchived 表示归档/下架)
)

func (s Status) String() string {
	// 定义状态名称映射数组
	// 索引对应 Status 的值
	statusNames := []string{
		"草稿",  // 0
		"审核中", // 1
		"已发布", // 2
		"已下线", // 3
	}

	// 边界检查，防止越界 panic
	if s >= 0 && int(s) < len(statusNames) {
		return statusNames[s]
	}
	return "未知状态"
}

//go:embed mappings/article_mapping.json
var articleMapping string //使用宏把json赋值到这里来

func (ArticleModel) Mapping() string {
	return articleMapping
}
func (ArticleModel) Index() string {
	return "article_index" //返回文章模型的索引名
}

// ArticleSearchModel 没有ES时的搜索降级搜索模型
type ArticleSearchModel struct {
	Model
	ArticleID uint         `gorm:"uniqueIndex" json:"articleID"`                      // 外键关联文章ID,唯一索引防止重复记录
	Article   ArticleModel `gorm:"foreignKey:ArticleID" json:"-"`                     //预备字段
	Title     string       `gorm:"size:32;index:idx_titleSearch" json:"title"`        // 文章标题，加索引
	Abstract  string       `gorm:"size:256;index:idx_abstractSearch" json:"abstract"` // 文章摘要，加索引
	TagList   string       `gorm:"size:256" json:"tagList"`                           // 标签JSON字符串,降级搜索按标签过滤用
}

// NewArticleSearchModel 由文章生成搜索记录
// 说明:标题/摘要按列宽截断,标签序列化为JSON字符串
func NewArticleSearchModel(article ArticleModel) ArticleSearchModel {
	tagJSON, err := json.Marshal(article.TagList)
	if err != nil || article.TagList == nil {
		tagJSON = []byte("[]")
	}
	return ArticleSearchModel{
		ArticleID: article.ID,
		Title:     truncateText(article.Title, 32),
		Abstract:  truncateText(article.Abstract, 256),
		TagList:   string(tagJSON),
	}
}

// truncateText 按字符数截断文本
// 参数:text - 原始文本
// 参数:maxLen - 最大字符数
// 返回:string - 截断后的文本
// 说明:按rune截断，避免中文等多字节字符被截断
func truncateText(text string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(text))
	if len(runes) <= maxLen {
		return string(runes)
	}
	return string(runes[:maxLen])
}

//文章创建时自动创建,并更新全文搜索记录
// 文章更新时,也会更新全文搜索记录
//删除时也会删除全文搜索记录

// 以下用于ES服务降级,当ES无法使用时,将使用数据库存储全文搜索记录

// AfterCreate 文章创建后写入全文搜索记录
// 参数:tx - GORM事务对象
// 返回:err - 错误信息
// 说明:仅发布状态的文章会被提取正文并保存到ArticleSearch中，用于降级搜索
func (this *ArticleModel) AfterCreate(tx *gorm.DB) (err error) {
	if this.Status != StatusPublished {
		return nil
	}
	//内容在创建时已经过滤了
	record := NewArticleSearchModel(*this)
	if err = tx.Create(&record).Error; err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}

// AfterDelete 文章删除后清理全文搜索记录
// 参数:tx - GORM事务对象
// 返回:err - 错误信息
// 说明:根据文章ID查找并删除相关的全文搜索记录
func (this *ArticleModel) AfterDelete(tx *gorm.DB) (err error) {
	var Result ArticleSearchModel
	tx.Delete(&Result, "article_id = ?", this.ID)
	return nil
}

// AfterUpdate 文章更新后刷新全文搜索记录
// 参数:tx - GORM事务对象
// 返回:err - 错误信息
// 说明:就地更新已有搜索记录(保留记录ID,避免编辑后搜索排序被顶到最前);
// 更新不会回写结构体,因此必须重新查询,避免用更新前的旧状态判断是否发布
func (this *ArticleModel) AfterUpdate(tx *gorm.DB) (err error) {
	//通过 Model(&ArticleModel{}).Where(...).Updates(...) 更新时,钩子接收者没有ID,跳过重建
	if this.ID == 0 {
		return nil
	}
	var current ArticleModel
	if err = tx.Select("id", "status", "title", "abstract", "tag_list").Take(&current, this.ID).Error; err != nil {
		return err
	}
	//未发布:删除搜索记录,避免已下线文章被搜到
	if current.Status != StatusPublished {
		return tx.Delete(&ArticleSearchModel{}, "article_id = ?", current.ID).Error
	}

	record := NewArticleSearchModel(current)
	result := tx.Model(&ArticleSearchModel{}).Where("article_id = ?", current.ID).
		Updates(map[string]any{
			"title":    record.Title,
			"abstract": record.Abstract,
			"tag_list": record.TagList,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		//内容没变化时RowsAffected可能为0,需确认记录是否真的不存在
		var count int64
		if err = tx.Model(&ArticleSearchModel{}).Where("article_id = ?", current.ID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return tx.Create(&record).Error
		}
	}
	return nil
}

// BeforeDelete 删除文章前清理关联数据
// 参数:tx - GORM事务对象
// 返回:err - 错误信息
// 说明:删除文章前，级联删除评论、点赞、收藏、置顶、浏览等关联记录
func (this *ArticleModel) BeforeDelete(tx *gorm.DB) (err error) {
	//使用批量DELETE,避免把热门文章的全部关联记录加载进内存
	//SkipHooks:避免UserArticleCollectModel等钩子在批量删除时被零值模型触发,污染Redis计数
	//NewDB:必须为每次级联删除开启全新语句,否则会继承外层ArticleModel删除语句的表名与主键条件
	//      (表现为 SQL 变成 DELETE FROM article_models WHERE article_id = ? AND id = ?,直接报错)
	batch := tx.Session(&gorm.Session{NewDB: true, SkipHooks: true})
	targets := []any{
		&CommentModel{},
		&ArticleDiggModel{},
		&UserArticleCollectModel{},
		&UserTopArticleModel{},
		&UserArticleHistoryModel{},
	}
	for _, target := range targets {
		if err = batch.Where("article_id = ?", this.ID).Delete(target).Error; err != nil {
			logrus.Errorf("删除文章 %d 的关联数据失败: %v", this.ID, err)
			return err
		}
	}
	logrus.Infof("已清理文章 %d 的关联数据", this.ID)
	return nil
}

// ArticleAddition 文章扩展附录表
// 通过 article_id 外键与文章一对一关联,存放管理员附加评论与AI点评(评级/摘要)
// AI 点评记录生成它的模型名(aiModel),便于追溯是哪个AI点评的文章
type ArticleAddition struct {
	Model
	ArticleID    uint   `gorm:"uniqueIndex" json:"articleID"`  // 关联文章ID,每篇文章一条
	AdminComment string `gorm:"size:1024" json:"adminComment"` // 管理员附加评论
	AIQuality    string `gorm:"size:255" json:"aiQuality"`     // AI生成的内容质量评级
	AIAbstract   string `gorm:"size:1024" json:"aiAbstract"`   // AI生成的内容摘要
	AIModel      string `gorm:"size:64" json:"aiModel"`        // 完成本次AI点评的模型名
}

// TableName 显式指定表名
func (ArticleAddition) TableName() string {
	return "article_additions"
}
