package ip

import (
	"StarDreamerCyberNook/global"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

// searchMu 保护 IPsearcher:ip2region 的 Searcher 非线程安全(内部 ioCount/缓冲会被并发改写)
var searchMu sync.Mutex

// GetIpAddr 根据IP地址获取地理位置信息
// 参数:ip - 要查询的IP地址字符串
// 返回:addr - 格式化后的地理位置字符串，格式为"省份·城市"或"国家·省份"等
// 说明:对于本地IP地址，返回"未知的本地ip"对于无效IP地址，返回"异常地址",对于格式异常的查询结果，返回"未知地址",优先显示省份和城市信息，其次是国家信息
func GetIpAddr(ip string) (addr string) {
	if HasLocalIPAddr(ip) {
		return "本地ip"
	}
	if global.IPsearcher == nil {
		return "未知地址"
	}

	parsed := net.ParseIP(ip)
	if parsed == nil {
		return "异常地址"
	}
	// 数据库为 IPv4 库,IPv6 地址无法查询
	if parsed.To4() == nil {
		return "未知地址"
	}

	//并发保护:Seek/Read 与内部计数非线程安全,加锁避免结果错乱
	searchMu.Lock()
	region, err := global.IPsearcher.Search(ip)
	searchMu.Unlock()
	if err != nil {
		logrus.Warnf("错误的ip地址 %s", err)
		return "异常地址"
	}
	if region == "" {
		return "未知地址"
	}
	_addrList := strings.Split(region, "|")

	// 兼容两种数据格式:
	//   5 段:国家|区域|省份|城市|运营商
	//   4 段:国家|省份|城市|运营商(xdb 新版默认格式)
	var country, province, city string
	switch len(_addrList) {
	case 5:
		country, province, city = _addrList[0], _addrList[2], _addrList[3]
	case 4:
		country, province, city = _addrList[0], _addrList[1], _addrList[2]
	default:
		// 数据库返回的格式异常，记录警告日志
		logrus.Warnf("异常的ip地址 %s, region=%q", ip, region)
		return "异常地址"
	}

	// 按照优先级格式化地址信息
	// 1. 优先显示省份和城市（当两者都有效时）
	if province != "0" && city != "0" {
		return fmt.Sprintf("%s·%s", province, city)
	}
	// 2. 其次显示国家和省份
	if country != "0" && province != "0" {
		return fmt.Sprintf("%s·%s", country, province)
	}
	// 3. 最后只显示国家
	if country != "0" {
		return country
	}
	// 4. 如果以上都无效，返回原始查询结果
	return region
}

func HasLocalIPAddr(ip string) bool {
	return HasLocalIP(net.ParseIP(ip))
}

// HasLocaLIP 检测 IP 地址是否是内网地址// 通过直接对比ip段范围效率更高
func HasLocalIP(ip net.IP) bool {
	if ip.IsLoopback() {
		return true
	}

	ip4 := ip.To4()
	if ip4 == nil {
		return false
	}

	return ip4[0] == 10 || // 10.0.0.0/8
		(ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) || // 172.16.0.0/12
		(ip4[0] == 169 && ip4[1] == 254) || // 169.254.0.0/16
		(ip4[0] == 192 && ip4[1] == 168) // 192.168.0.0/16
}
