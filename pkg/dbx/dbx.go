// Package dbx 按 driver/dsn 统一打开数据库连接池。
//
// 宿主模式下,未启用独立库的微服务复用宿主的 global.DB(不各自建池);
// 仅当微服务显式启用独立库(Config.Standalone)时才调用本包自建连接池。
package dbx

import (
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// 连接池默认参数(与 core.InitDB 保持一致)。
const (
	maxIdleConns = 10
	maxOpenConns = 100
	connMaxLife  = time.Hour
)

// Open 按 driver(mysql|postgres|sqlite)与 dsn 打开 gorm,并套用统一连接池参数。
func Open(driver, dsn string) (*gorm.DB, error) {
	var dial gorm.Dialector
	switch driver {
	case "postgres":
		dial = postgres.Open(dsn)
	case "sqlite":
		dial = sqlite.Open(dsn)
	default:
		dial = mysql.Open(dsn)
	}
	db, err := gorm.Open(dial, &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		return nil, err
	}
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxIdleConns(maxIdleConns)
		sqlDB.SetMaxOpenConns(maxOpenConns)
		sqlDB.SetConnMaxLifetime(connMaxLife)
	}
	return db, nil
}
