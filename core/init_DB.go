// core/init_db.go
package core

import (
	"StarDreamerCyberNook/conf"
	"StarDreamerCyberNook/global"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// InitDB 初始化数据库连接
// 参数: 无
// 返回: *gorm.DB - 数据库连接实例
// 说明: 配置已彻底拆分为写库列表(dbWrite)与读库列表(dbRead);
// 读库列表为空时,读请求也会落到写库上
func InitDB() *gorm.DB {
	writeList := global.Config.DBWrite
	readList := global.Config.DBRead

	if len(writeList) == 0 {
		logrus.Fatalf("未配置写库(dbWrite)")
	}

	// 所有库的数据库类型必须一致
	all := make([]conf.DB, 0, len(writeList)+len(readList))
	all = append(all, writeList...)
	all = append(all, readList...)
	for i := 1; i < len(all); i++ {
		if all[i].SqlName != all[0].SqlName {
			logrus.Fatalf("数据库配置错误: 模式不一致, %s != %s", all[i].SqlName, all[0].SqlName)
		}
	}

	// 连接写库,第一个写库作为主连接
	sources := make([]gorm.Dialector, 0, len(writeList))
	for _, v := range writeList {
		logrus.Infof("写库配置: 模式=%s, 主机=%s, 端口=%d, 数据库=%s", v.SqlName, v.Host, v.Port, v.DBName)
		sources = append(sources, v.DSN())
	}
	db, err := gorm.Open(sources[0], &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true, // 不生成外键约束
	})
	if err != nil {
		logrus.Fatalf("数据库连接失败 %s", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		logrus.Fatalf("获取数据库连接池失败 %s", err)
	}
	sqlDB.SetMaxIdleConns(10)           // 最大空闲连接数
	sqlDB.SetMaxOpenConns(100)          // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生命周期
	logrus.Infof("数据库连接成功, 写库 %d 个, 读库 %d 个", len(writeList), len(readList))

	if len(readList) > 0 {
		// 注册读写分离:写请求走 Sources,读请求走 Replicas
		replicas := make([]gorm.Dialector, 0, len(readList))
		for _, v := range readList {
			logrus.Infof("读库配置: 模式=%s, 主机=%s, 端口=%d, 数据库=%s", v.SqlName, v.Host, v.Port, v.DBName)
			replicas = append(replicas, v.DSN())
		}
		err = db.Use(dbresolver.Register(dbresolver.Config{
			Sources:  sources,  // 写库
			Replicas: replicas, // 读库
			// 负载均衡策略：随机选择 replica
			Policy: dbresolver.RandomPolicy{},
		}))
		if err != nil {
			logrus.Fatalf("读写配置出错: %s", err)
		}
	}

	if global.Config.System.RunMode == "debug" {
		db = db.Debug()
		logrus.Debug("数据库调试模式已开启")
	}
	return db
}
