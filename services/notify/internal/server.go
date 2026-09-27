package internal

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	notifyv1 "StarDreamerCyberNook/gen/notify/v1"

	"github.com/jordan-wright/email"
)

const smtpTimeout = 15 * time.Second

const codeBoxStyle = "background:#f5f7ff; border:1px dashed #a5b4fc; padding:20px; border-radius:10px; text-align:center; font-size:32px; letter-spacing:10px; color:#4f46e5; font-weight:bold; font-family:'SFMono-Regular',Consolas,'Liberation Mono',Menlo,monospace;"

// Config 是 notify 服务的 SMTP 与站点配置(来自配置文件,环境变量可选覆盖)。
type Config struct {
	Domain       string
	Port         int
	SendEmail    string
	AuthCode     string
	SendNickname string
	SSL          bool
	TLS          bool
	SiteTitle    string
}

// Server 实现 NotifyService。
type Server struct {
	notifyv1.UnimplementedNotifyServiceServer
	cfg Config
}

// NewServer 创建服务。
func NewServer(cfg Config) *Server { return &Server{cfg: cfg} }

func wrapEmailHTML(body string) string {
	return fmt.Sprintf(`<div style="background:#eef1f6; padding:32px 16px; font-family:-apple-system,BlinkMacSystemFont,'Segoe UI','PingFang SC','Microsoft YaHei',sans-serif;">
  <div style="max-width:560px; margin:0 auto; background:#ffffff; border-radius:14px; padding:28px 32px; box-shadow:0 6px 24px rgba(15,23,42,0.08); color:#334155; font-size:15px; line-height:1.7;">
%s
  </div>
</div>`, body)
}

// SendRegister 发送注册验证码邮件。
func (s *Server) SendRegister(_ context.Context, req *notifyv1.SendCodeRequest) (*notifyv1.SendEmailReply, error) {
	subject := fmt.Sprintf("%s - 注册验证码", s.cfg.SendNickname)
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
		req.GetTarget(), s.cfg.SiteTitle, codeBoxStyle, req.GetCode(), s.cfg.SendEmail, s.cfg.SendEmail, s.cfg.SiteTitle)
	return s.send(req.GetTarget(), subject, wrapEmailHTML(body))
}

// SendForgetPwd 发送重置密码验证码。
func (s *Server) SendForgetPwd(_ context.Context, req *notifyv1.SendCodeRequest) (*notifyv1.SendEmailReply, error) {
	subject := fmt.Sprintf("%s - 密码重置验证码", s.cfg.SendNickname)
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
		req.GetTarget(), codeBoxStyle, req.GetCode(), s.cfg.SiteTitle)
	return s.send(req.GetTarget(), subject, wrapEmailHTML(body))
}

// SendResetEmail 发送重置邮箱验证码。
func (s *Server) SendResetEmail(_ context.Context, req *notifyv1.SendCodeRequest) (*notifyv1.SendEmailReply, error) {
	subject := fmt.Sprintf("%s - 邮箱重置验证码", s.cfg.SendNickname)
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
		req.GetTarget(), codeBoxStyle, req.GetCode(), s.cfg.SiteTitle)
	return s.send(req.GetTarget(), subject, wrapEmailHTML(body))
}

// SendEmail 通用发送。
func (s *Server) SendEmail(_ context.Context, req *notifyv1.SendEmailRequest) (*notifyv1.SendEmailReply, error) {
	return s.send(req.GetTo(), req.GetSubject(), req.GetHtml())
}

// send 底层 SMTP 发送(逻辑与原 email_service.SendEmail 一致)。
func (s *Server) send(to, subject, html string) (*notifyv1.SendEmailReply, error) {
	em := s.cfg
	if em.Domain == "" || em.SendEmail == "" || em.Port == 0 {
		return nil, fmt.Errorf("邮件服务未配置: domain/sendEmail/port 不能为空")
	}
	if to == "" {
		return nil, fmt.Errorf("收件人不能为空")
	}

	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", em.SendNickname, em.SendEmail)
	e.To = []string{to}
	e.Subject = subject
	e.HTML = []byte(html)
	e.Text = []byte("如果html文本不可视,可查看此处" + stripHTML(html))

	sender, err := mail.ParseAddress(em.SendEmail)
	if err != nil {
		return nil, fmt.Errorf("发件人地址无效: %w", err)
	}
	raw, err := e.Bytes()
	if err != nil {
		return nil, fmt.Errorf("构建邮件内容失败: %w", err)
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
		return nil, fmt.Errorf("连接邮件服务器失败: %w", err)
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(smtpTimeout)); err != nil {
		return nil, fmt.Errorf("设置邮件超时失败: %w", err)
	}

	client, err := smtp.NewClient(conn, em.Domain)
	if err != nil {
		return nil, fmt.Errorf("初始化SMTP客户端失败: %w", err)
	}
	defer client.Close()

	if em.TLS && !em.SSL {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return nil, fmt.Errorf("邮件服务器不支持 STARTTLS")
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return nil, fmt.Errorf("STARTTLS 失败: %w", err)
		}
	}

	if em.AuthCode != "" {
		if ok, _ := client.Extension("AUTH"); !ok {
			return nil, fmt.Errorf("邮件服务器不支持认证")
		}
		if err := client.Auth(smtp.PlainAuth("", em.SendEmail, em.AuthCode, em.Domain)); err != nil {
			return nil, fmt.Errorf("SMTP 认证失败: %w", err)
		}
	}

	if err := client.Mail(sender.Address); err != nil {
		return nil, fmt.Errorf("设置发件人失败: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return nil, fmt.Errorf("设置收件人失败: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return nil, fmt.Errorf("准备发送邮件内容失败: %w", err)
	}
	if _, err := w.Write(raw); err != nil {
		return nil, fmt.Errorf("写入邮件内容失败: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("提交邮件内容失败: %w", err)
	}
	if err := client.Quit(); err != nil {
		return nil, fmt.Errorf("结束邮件会话失败: %w", err)
	}
	return &notifyv1.SendEmailReply{}, nil
}

func stripHTML(html string) string {
	result := strings.ReplaceAll(html, "<br/>", "\n")
	result = strings.ReplaceAll(result, "<br>", "\n")
	result = strings.ReplaceAll(result, "<div>", "")
	result = strings.ReplaceAll(result, "</div>", "")
	result = strings.ReplaceAll(result, "<b>", "")
	result = strings.ReplaceAll(result, "</b>", "")
	return result
}
