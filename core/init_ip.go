// core/init_ip_db.go
package core

import (
	"os"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/sirupsen/logrus"
)

// searcher 是IP地址数据库查询器实例，用于IP地址到地理位置的查询

// InitIPDB 初始化IP地址数据库
// 将 ip2region.xdb 一次性读入内存并构建基于内容缓冲的查询器。
// 说明:ip2region 的 Searcher 基于文件 Seek/Read 实现,并非线程安全;
// 使用内存缓冲可避免并发读取同一文件句柄导致的错乱(查询处仍需加锁,见 utils/ip)。
func InitIPDB() *xdb.Searcher {
	const dbPath = "init/ip2region.xdb"

	buff, err := os.ReadFile(dbPath)
	if err != nil {
		logrus.Fatalf("ip地址数据库读取失败 %s", err)
		return nil
	}

	searcher, err := xdb.NewWithBuffer(xdb.IPv4, buff)
	if err != nil {
		logrus.Fatalf("ip地址数据库加载失败 %s", err)
		return nil
	}
	return searcher
}
