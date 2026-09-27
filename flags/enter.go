package flags

import (
	"fmt"
	"strings"

	"StarDreamerCyberNook/conf"
)

// FileOpt 承载"带值选项"的参数:选项被使用时指针非 nil。
type FileOpt struct{ Path string }

// AccountOpt 承载"引导式创建账号"过程中收集到的字段。
type AccountOpt struct {
	UserName string
	Password string
	Email    string
	NickName string
	Role     int
}

// Option 每个选项 = (启用标记 / 子结构体指针):
// 使用某选项时,对应 bool 置 true 或指针指向有效实例;未使用则为 false / nil。
type Option struct {
	File    *FileOpt    // -f <path>        配置文件
	Account *AccountOpt // -create-user     引导式创建账号
	DB      bool        // -db              迁移数据库
	ES      bool        // -es              建立 ES 索引
	Search  bool        // -search          重建数据库搜索表
	Version bool        // -v               版本
}

// Parse 解析命令行参数(通常传 os.Args[1:]),返回一个可用的 Option。
// 支持 -k v、-k=v、--k v、--k=v。
func Parse(args []string) Option {
	var o Option
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if len(arg) < 2 || arg[0] != '-' {
			continue
		}
		key := strings.TrimLeft(arg, "-")
		val, hasVal := "", false
		if eq := strings.IndexByte(key, '='); eq >= 0 {
			val, key, hasVal = key[eq+1:], key[:eq], true
		}
		value := func() string {
			if hasVal {
				return val
			}
			if i+1 < len(args) {
				i++
				return args[i]
			}
			return ""
		}
		switch key {
		case "f", "file":
			o.File = &FileOpt{Path: value()}
		case "create-user", "cu":
			o.Account = &AccountOpt{}
		case "db":
			o.DB = true
		case "es":
			o.ES = true
		case "search":
			o.Search = true
		case "v", "version":
			o.Version = true
		}
	}
	return o
}

// ConfigPath 当前配置文件路径(未指定时默认 setting.yaml)。
func (o Option) ConfigPath() string {
	if o.File != nil && o.File.Path != "" {
		return o.File.Path
	}
	return "setting.yaml"
}

// Run 消费 Option 执行命令行附加功能(不调度平台;博客默认开启)。
// 各功能相互独立,顺序执行;失败即返回错误。
func Run(o Option) (err error) {
	if o.Version {
		fmt.Printf("当前版本为: %s\n", conf.Version)
	}
	if o.DB {
		FlagDB()
	}
	if o.ES {
		EsIndex()
	}
	if o.Search {
		FlagSearch()
	}
	if o.Account != nil {
		CreateUser(o.Account)
	}
	return nil
}
