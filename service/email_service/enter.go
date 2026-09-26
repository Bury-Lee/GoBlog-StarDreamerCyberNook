package email_service

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"StarDreamerCyberNook/global"

	"github.com/jordan-wright/email"
)

// smtpTimeout 单次邮件发送的整体超时时间,避免服务器无响应时请求被永久挂起
const smtpTimeout = 15 * time.Second

// codeBoxStyle 验证码展示框样式(仅样式美化,不改动文案)
const codeBoxStyle = "background:#f5f7ff; border:1px dashed #a5b4fc; padding:20px; border-radius:10px; text-align:center; font-size:32px; letter-spacing:10px; color:#4f46e5; font-weight:bold; font-family:'SFMono-Regular',Consolas,'Liberation Mono',Menlo,monospace;"

// wrapEmailHTML 给邮件正文套上统一的美化容器(仅调整样式,不改变正文文字)
func wrapEmailHTML(body string) string {
	return fmt.Sprintf(`<div style="background:#eef1f6; padding:32px 16px; font-family:-apple-system,BlinkMacSystemFont,'Segoe UI','PingFang SC','Microsoft YaHei',sans-serif;">
  <div style="max-width:560px; margin:0 auto; background:#ffffff; border-radius:14px; padding:28px 32px; box-shadow:0 6px 24px rgba(15,23,42,0.08); color:#334155; font-size:15px; line-height:1.7;">
%s
  </div>
</div>`, body)
}

// SendRegister 发送注册验证码邮件
func SendRegister(target string, code string) error {
	em := global.Config.Email
	subject := fmt.Sprintf("%s - 注册验证码", em.SendNickname)

	body := fmt.Sprintf(`亲爱的用户 <b>%s</b> 大人，<br/><br/>
来自遥远世界的讯息传来：<br/>
🎉 恭喜您触发了 <b>%s</b> 的注册事件！<br/>
<br/>
您的验证码已经生成完毕，正在通过网络协议以光速奔向您的收件箱：<br/>
<br/>
<div style="%s">
    %s
</div>
<br/>
（2分钟后自动过期）<br/>
<br/>
• 请勿将验证码泄露给其他人<br/>
• 如果这不是您本人操作，可能是您的邮箱被平行世界的您借用了<br/>
• 本邮件由程序自动发送，没有人类受到伤害（除了写代码的笔者）<br/>
<br/>
遇到 Bug 或想吐槽？欢迎前来反馈~<br/>
邮箱：<a href="mailto:%s" style="color:#4f46e5; text-decoration:none;">%s</a><br/>
<br/>
<b>%s</b> 敬上`,
		target,
		global.Config.Site.Project.Title,
		codeBoxStyle,
		code,
		em.SendEmail,
		em.SendEmail,
		global.Config.Site.Project.Title,
	)

	return SendEmail(target, subject, wrapEmailHTML(body))
}

// SendForgetPwd 发送重置密码验证码
func SendForgetPwd(target string, code string) error {
	em := global.Config.Email
	subject := fmt.Sprintf("%s - 密码重置验证码", em.SendNickname)

	body := fmt.Sprintf(`亲爱的用户 <b>%s</b> 大人，<br/><br/>
您正在申请重置密码，验证码如下：<br/>
<br/>
<div style="%s">
    %s
</div>
<br/>
（2分钟后自动过期）<br/>
<br/>
• 请勿将验证码泄露给其他人<br/>
• 如果这不是您本人操作，请立即修改密码<br/>
<br/>
<b>%s</b> 敬上`,
		target,
		codeBoxStyle,
		code,
		global.Config.Site.Project.Title,
	)

	return SendEmail(target, subject, wrapEmailHTML(body))
}

// SendResetEmail 发送重置邮箱验证码
func SendResetEmail(target string, code string) error {
	em := global.Config.Email
	subject := fmt.Sprintf("%s - 邮箱重置验证码", em.SendNickname)

	body := fmt.Sprintf(`亲爱的用户 <b>%s</b> 大人，<br/><br/>
您正在申请重置邮箱，验证码如下：<br/>
<br/>
<div style="%s">
    %s
</div>
<br/>
（2分钟后自动过期）<br/>
<br/>
• 请勿将验证码泄露给其他人<br/>
• 如果这不是您本人操作，请立即修改密码<br/>
<br/>
<b>%s</b> 敬上`,
		target,
		codeBoxStyle,
		code,
		global.Config.Site.Project.Title,
	)

	return SendEmail(target, subject, wrapEmailHTML(body))
}

// SendEmail 发送邮件的通用函数
func SendEmail(to, subject, text string) error {
	em := global.Config.Email
	if em.Domain == "" || em.SendEmail == "" || em.Port == 0 {
		return fmt.Errorf("邮件服务未配置: domain/sendEmail/port 不能为空")
	}
	if to == "" {
		return fmt.Errorf("收件人不能为空")
	}

	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", em.SendNickname, em.SendEmail)
	e.To = []string{to}
	e.Subject = subject
	e.HTML = []byte(text)
	// 添加文本版本作为后备（提高兼容性）
	e.Text = []byte("如果html文本不可视,可查看此处" + stripHTML(text))

	sender, err := mail.ParseAddress(em.SendEmail)
	if err != nil {
		return fmt.Errorf("发件人地址无效: %w", err)
	}
	raw, err := e.Bytes()
	if err != nil {
		return fmt.Errorf("构建邮件内容失败: %w", err)
	}

	addr := net.JoinHostPort(em.Domain, strconv.Itoa(em.Port))
	tlsConfig := &tls.Config{ServerName: em.Domain}

	dialer := &net.Dialer{Timeout: smtpTimeout}
	var conn net.Conn
	if !em.SSL {
		conn, err = dialer.Dial("tcp", addr)
	} else {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	}
	if err != nil {
		return fmt.Errorf("连接邮件服务器失败: %w", err)
	}
	defer conn.Close()

	// 给整个 SMTP 会话设置读写截止时间,端口/加密方式不匹配时不会无限阻塞
	if err := conn.SetDeadline(time.Now().Add(smtpTimeout)); err != nil {
		return fmt.Errorf("设置邮件超时失败: %w", err)
	}

	client, err := smtp.NewClient(conn, em.Domain)
	if err != nil {
		return fmt.Errorf("初始化SMTP客户端失败: %w", err)
	}
	defer client.Close()

	// 显式 TLS:在明文连接上协商 STARTTLS(SSL 已用隐式 TLS,无需再次升级)
	if em.TLS && !em.SSL {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return fmt.Errorf("邮件服务器不支持 STARTTLS")
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("STARTTLS 失败: %w", err)
		}
	}

	if em.AuthCode != "" {
		if ok, _ := client.Extension("AUTH"); !ok {
			return fmt.Errorf("邮件服务器不支持认证")
		}
		if err := client.Auth(smtp.PlainAuth("", em.SendEmail, em.AuthCode, em.Domain)); err != nil {
			return fmt.Errorf("SMTP 认证失败: %w", err)
		}
	}

	if err := client.Mail(sender.Address); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("设置收件人失败: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("准备发送邮件内容失败: %w", err)
	}
	if _, err := w.Write(raw); err != nil {
		return fmt.Errorf("写入邮件内容失败: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("提交邮件内容失败: %w", err)
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("结束邮件会话失败: %w", err)
	}
	return nil
}

// stripHTML 简单的 HTML 标签去除函数（用于生成纯文本版本）
func stripHTML(html string) string {
	// 简单的标签替换，生产环境建议使用 bluemonday 或 html2text 库
	result := strings.ReplaceAll(html, "<br/>", "\n")
	result = strings.ReplaceAll(result, "<br>", "\n")
	result = strings.ReplaceAll(result, "<div>", "")
	result = strings.ReplaceAll(result, "</div>", "")
	result = strings.ReplaceAll(result, "<b>", "")
	result = strings.ReplaceAll(result, "</b>", "")
	// 移除其他 HTML 标签的简单正则或字符串操作
	return result
}
