package flags

import (
	"fmt"
	"os"
	"time"

	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	Hash "StarDreamerCyberNook/utils/hash"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
)

// CreateUser 引导式创建账号:交互收集字段到 opt,校验后写库。
func CreateUser(opt *AccountOpt) {
	fmt.Println("选择角色     1 管理员   2 普通用户   3 游客")
	if _, err := fmt.Scan(&opt.Role); err != nil {
		logrus.Errorf("读取角色失败 %s", err)
		return
	}
	if !(opt.Role == 1 || opt.Role == 2 || opt.Role == 3) {
		logrus.Errorf("角色非法")
		return
	}

	fmt.Println("输入用户名:")
	fmt.Scan(&opt.UserName)
	var model models.UserModel
	if global.DB.Take(&model, "user_name = ?", opt.UserName).Error == nil {
		logrus.Errorf("用户名已存在")
		return
	}

	fmt.Println("输入邮箱:")
	fmt.Scan(&opt.Email)
	if global.DB.Take(&model, "email = ?", opt.Email).Error == nil {
		logrus.Errorf("邮箱已存在")
		return
	}

	fmt.Println("输入密码:")
	pwd, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("读取密码时出错:", err)
		return
	}
	fmt.Println("请再次输入密码:")
	rePwd, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("读取密码时出错:", err)
		return
	}
	if string(pwd) != string(rePwd) {
		fmt.Println("两次密码不一致")
		return
	}
	opt.Password = string(pwd)

	fmt.Println("输入昵称(为空自动默认):")
	fmt.Scan(&opt.NickName)
	if opt.NickName == "" {
		opt.NickName = "默认"
	}

	hashPwd, _ := Hash.HashPassword(opt.Password)
	if err := global.DB.Create(&models.UserModel{
		UserName:       opt.UserName,
		NickName:       opt.NickName,
		RegisterSource: enum.RegisterTerminal,
		Password:       hashPwd,
		Email:          opt.Email,
		Role:           enum.RoleType(opt.Role),
		LastLoginTime:  time.Now(),
	}).Error; err != nil {
		fmt.Println("创建用户失败", err)
		return
	}
	logrus.Infof("创建用户成功")
}
