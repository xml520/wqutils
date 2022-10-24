package appleTools

import (
	"errors"
	"github.com/jhillyerd/enmime"
	"github.com/xml520/go-smtp"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// The Backend implements SMTP server methods.
type backendV2 struct {
	Hook  MailHookV2
	deBug bool
}
type MailHookV2 interface {
	TeamInvite(*EmailTeamInviteType)
	NoType(*EmailNoType)
}

func (bkd *backendV2) NewSession(_ *smtp.Conn) (smtp.Session, error) {
	return &sessionV2{Hook: bkd.Hook, debug: bkd.deBug}, nil
}

type sessionV2 struct {
	To    string
	Hook  MailHookV2
	debug bool
}

func (s *sessionV2) AuthPlain(_, _ string) error {
	return nil
}
func (s *sessionV2) Mail(from string, _ *smtp.MailOptions) error {
	if s.debug {
		log.Println("Mail from:", from)
	} else {
		if strings.Index(from, appleDomain) == -1 {
			return errors.New("no authorized")
		}
	}
	return nil
}
func (s *sessionV2) Rcpt(to string) error {
	if s.debug {
		log.Println("Rcpt to:", to)
	}
	s.To = to
	return nil
}
func (s *sessionV2) Data(r io.Reader) (err error) {
	var c *enmime.Envelope
	var m *MailContent
	c, err = enmime.ReadEnvelope(r)
	if err != nil {
		return err
	}
	m = &MailContent{ParserContent: c}
	m.Subject = strings.TrimSpace(m.ParserContent.GetHeader("Subject"))
	m.To = strings.TrimSpace(s.To)
	return handleTypeV2(m, s.Hook, s.debug)
}
func (s *sessionV2) Reset() {}
func (s *sessionV2) Logout() error {
	return nil
}

// MailServerListenV2 MailServerListen 开启邮件服务
func MailServerListenV2(addr string, hook MailHookV2, debug bool) error {
	s := smtp.NewServer(&backendV2{hook, debug})
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

func handleTypeV2(m *MailContent, hook MailHookV2, debug bool) error {
	switch {
	case m.Subject == emailPrivateInviteSubject:
		hook.TeamInvite(&EmailTeamInviteType{
			Key:         m.MiddleStr("activation_ds?key=", "\">Accept invitation<"),
			MailContent: m,
		})

	case m.Subject == emailCompanyInviteSubject:
		hook.TeamInvite(&EmailTeamInviteType{
			Key:         m.MiddleStr("activation_ds?key=", "&provider"),
			MailContent: m,
		})
	default:
		hook.NoType(&EmailNoType{m})
		if debug {
			os.WriteFile("../tmp/"+strconv.FormatInt(time.Now().Unix(), 10)+".html", []byte(m.ParserContent.Text), 0644)
		}
		return errors.New("not")
		//hook.NoType(&EmailNoType{m})
		//os.WriteFile("../tmp/tmp.html", []byte(m.ParserContent.Text), 0755)
	}
	return nil
}
