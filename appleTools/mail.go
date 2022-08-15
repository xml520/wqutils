package appleTools

import (
	"errors"
	"github.com/jhillyerd/enmime"
	"github.com/xml520/go-smtp"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	appleDomain               = "apple.com"
	emailCompanyInviteSubject = "You're invited to join a development team." // 公司邀请
	emailPrivateInviteSubject = "You've been invited to App Store Connect."  // 个人邀请
	verifyEmailSubject        = "验证您的 Apple ID 电子邮件地址"
	verifyAppleIDSubject      = "Verify your Apple ID."
)

// EmailNoType 未匹配到类型
type EmailNoType struct {
	*MailContent
}

// EmailVerifyAppleIDType 测试员邮箱验证
type EmailVerifyAppleIDType struct {
	*MailContent
}

// EmailVerifyEmailType 注册时候验证Apple 邮箱
type EmailVerifyEmailType struct {
	Code string
	*MailContent
}

// EmailBuildFailedType 构建失败类型
type EmailBuildFailedType struct {
	ID uint64 // appid
	*MailContent
}

// EmailTeamInviteType  苹果团队邀请类型
type EmailTeamInviteType struct {
	Key string
	*MailContent
}

// GetID 尝试获取开发者账号本地数据库ID 邀请人的名字格式必须为 id.$id.id
func (e *EmailTeamInviteType) GetID() (int, error) {
	idStr := e.MiddleStr("<p>Dear id.", ".id")
	return strconv.Atoi(idStr)
}

// EmailTestflightLinkType tf链接类型
type EmailTestflightLinkType struct {
	AppID   uint64 // 应用ID
	LinkKey string // 下载链接key
	*MailContent
}

type MailContent struct {
	Subject       string
	To            string
	ParserContent *enmime.Envelope
}
type MailHook interface {
	TestflightLink(*EmailTestflightLinkType)
	BuildFailed(*EmailBuildFailedType)
	VerifyEmail(*EmailVerifyEmailType)
	VerifyAppleID(*EmailVerifyAppleIDType)
	NoType(*EmailNoType)
	TeamInvite(*EmailTeamInviteType)
}

// MiddleStr 取中间文本
func (m *MailContent) MiddleStr(left string, right string) string {
	li := strings.Index(m.ParserContent.HTML, left)
	if li == -1 {
		return ""
	}
	li = li + len(left)
	ri := strings.Index(m.ParserContent.HTML[li:], right)
	if ri == -1 {
		return ""
	}
	return strings.TrimSpace(m.ParserContent.HTML[li : li+ri])
}

// The Backend implements SMTP server methods.
type backend struct {
	hook MailHook
}

func (bkd *backend) NewSession(_ *smtp.Conn) (smtp.Session, error) {
	return &session{bkd.hook, ""}, nil
}

type session struct {
	hook MailHook
	To   string
}

func (s *session) AuthPlain(_, _ string) error {
	return nil
}
func (s *session) Mail(from string, _ *smtp.MailOptions) error {
	log.Println("Mail from:", from)
	if strings.Index(from, appleDomain) == -1 {
		return errors.New("no authorized")
	}
	return nil
}
func (s *session) Rcpt(to string) error {
	log.Println("Rcpt to:", to)
	s.To = to
	return nil
}
func (s *session) Data(r io.Reader) (err error) {
	var c *enmime.Envelope
	var m *MailContent
	c, err = enmime.ReadEnvelope(r)
	if err != nil {
		return err
	}
	m = &MailContent{ParserContent: c}
	m.Subject = strings.TrimSpace(m.ParserContent.GetHeader("Subject"))
	m.To = strings.TrimSpace(s.To)
	return handleType(m, s.hook)
}
func (s *session) Reset() {}
func (s *session) Logout() error {
	return nil
}

// MailServerListen 开启邮件服务
func MailServerListen(addr string, hook MailHook) error {

	s := smtp.NewServer(&backend{hook})
	s.Addr = addr
	s.Domain = "localhost"
	s.ReadTimeout = 10 * time.Second
	s.WriteTimeout = 10 * time.Second
	s.MaxMessageBytes = 1024 * 1024
	s.MaxRecipients = 50
	s.AllowInsecureAuth = true
	log.Println("绑定25端口", s.Addr)
	return s.ListenAndServe()
}

func handleType(m *MailContent, hook MailHook) error {
	var key string
	if key = m.MiddleStr("v1/invite/", "?ct="); key != "" {
		l := &EmailTestflightLinkType{LinkKey: key, MailContent: m}
		l.AppID, _ = strconv.ParseUint(key[len(key)-8:], 16, 0)
		hook.TestflightLink(l)
		return nil
	}
	switch {
	case m.Subject == emailPrivateInviteSubject:
		//fmt.Println("邀请链接", m.MiddleStr("activation_ds?key=", "\">Accept invitation<"))
		hook.TeamInvite(&EmailTeamInviteType{
			Key:         m.MiddleStr("activation_ds?key=", "\">Accept invitation<"),
			MailContent: m,
		})

	case m.Subject == emailCompanyInviteSubject:
		hook.TeamInvite(&EmailTeamInviteType{
			Key:         m.MiddleStr("activation_ds?key=", "&provider"),
			MailContent: m,
		})
	case m.Subject == verifyAppleIDSubject:
		hook.VerifyAppleID(&EmailVerifyAppleIDType{m})
	case strings.Index(m.Subject, "has one or more issues") != -1:

		appid, err := strconv.Atoi(m.MiddleStr("Apple ID: ", " Version"))
		if err == nil {
			hook.BuildFailed(&EmailBuildFailedType{
				ID:          uint64(appid),
				MailContent: m,
			})
		}
	case m.Subject == verifyEmailSubject:
		var s, _ = regexp.Compile("[0-9]{6}")
		hook.VerifyEmail(&EmailVerifyEmailType{Code: s.FindString(m.ParserContent.Text), MailContent: m})

	default:
		hook.NoType(&EmailNoType{m})
		//os.WriteFile("../tmp/tmp.html", []byte(m.ParserContent.Text), 0755)
	}
	return nil
}
