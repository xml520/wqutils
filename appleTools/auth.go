package appleTools

import (
	"errors"
	"fmt"
	"github.com/xml520/wqutils/httpclient"
	"net/http"
	"strings"
)

const (
	authBaseUrl                      = `https://idmsa.apple.com/appleauth/auth`
	jsonContentType                  = `application/json`
	appleAuthXAppleWidgetKeyAppStore = `e0b80c3bf78523bfe80974d320935bfa30add02e1bff88ec2166c6bd5a706c42`
	language                         = `zh-CN,zh;q=0.9,ga;q=0.8,et;q=0.7`
)

var (
	authClient   *httpclient.HttpClient
	AuthError409 = errors.New("需要双重验证")
	AuthError412 = errors.New("未接受隐私协议")
	AuthError503 = errors.New("您的登录太频繁，请稍等一分钟再试")
)

func init() {
	authClient = httpclient.NewHttpClient().Defaults(httpclient.Map{
		"Accept-Language":    language,
		"X-Apple-Widget-Key": appleAuthXAppleWidgetKeyAppStore,
		"Accept":             jsonContentType,
		httpclient.OPT_AFTER_REQUEST_FUNC: func(res *httpclient.Response) error {
			if res.StatusCode < 299 {
				return nil
			}
			switch res.StatusCode {
			case 409:
				return AuthError409
			case 412:
				return AuthError412
			case 503:
				return AuthError503
			default:
				if msg := res.ToJson("serviceErrors.0.message").String(); msg != "" {
					return errors.New(msg)
				}
				if msg := res.ToJson("errorMessage").String(); msg != "" {
					return errors.New(msg)
				}
				return errors.New(fmt.Sprintf("%s 未知错误 状态码：%v", res.Request.URL.String(), res.Status))
			}
		},
		httpclient.OPT_TIMEOUT: 30,
	})
}

type Auth struct {
	Account  string `json:"account" gorm:"uniqueIndex;comment:账号名"`
	Password string `json:"password" gorm:"comment:密码"`
	Web
	proxy string
}
type AuthSession struct {
	Auth         *Auth             `json:"auth"`
	Header       map[string]string `json:"header"`
	Mobiles      []*authMobile     `json:"mobiles"`
	SelectMobile *authMobile       `json:"select_mobile"`
}
type authMobile struct {
	ID     int    `json:"id"`
	Mode   string `json:"mode"`
	Mobile string `json:"mobile"`
}

// SetProxy 设置代理IP
func (a *Auth) SetProxy(u string) {
	a.proxy = u
}

// SignIn 登录
func (a *Auth) SignIn() (session *AuthSession, err error) {
	var res *httpclient.Response
	session = &AuthSession{Auth: a}
	_url := authBaseUrl + `/signin?isRememberMeEnabled=true`
	_data := map[string]any{
		"accountName": a.Account,
		"password":    a.Password,
		"rememberMe":  true,
	}
	res, err = session.http().PostJson(_url, _data)
	if err != nil {
		switch err {
		case AuthError409:
			session.extractHeader(res)
			if err1 := session.extractMobile(); err1 != nil {
				err = fmt.Errorf(err.Error()+" %s", err1)
			}
			return
		case AuthError412:
			session.extractHeader(res, "X-Apple-Repair-Session-Token")
			if err = session.accept(); err != nil {
				return
			}
		default:
			return
		}
	}
	a.setHttpCookie(res.Cookies())
	return
}

// CheckCode 验证
func (a *AuthSession) CheckCode(code string) error {
	var (
		_url  string
		_data map[string]any
	)
	if a.SelectMobile.Mode == "sms" {
		_url = authBaseUrl + `/verify/phone/securitycode`
		_data = map[string]any{
			"phoneNumber": map[string]int{
				"id": a.SelectMobile.ID,
			},
			"securityCode": map[string]string{
				"code": code,
			},
			"mode": "sms",
		}
	} else {
		_url = authBaseUrl + `/verify/trusteddevice/securitycode`
		_data = map[string]any{
			"securityCode": map[string]string{
				"code": code,
			},
		}
	}
	res, err := a.http().PostJson(_url, _data)
	if err != nil {
		return err
	}
	a.extractHeader(res)
	return a.trustCookie()
}

// SendSMS 重新发送验证码
func (a *AuthSession) SendSMS() error {
	res, err := a.http().PutJson(authBaseUrl+"verify/phone", map[string]interface{}{
		"phoneNumber": map[string]int{"id": a.SelectMobile.ID},
		"mode":        "sms",
	})
	if err != nil {
		return err
	}
	a.SelectMobile.Mode = "sms"
	a.extractHeader(res)
	return err
}
func (a *AuthSession) trustCookie() error {
	if res, err := a.http().Get(authBaseUrl+"/2sv/trust", nil); err != nil {
		return err
	} else {
		a.Auth.setHttpCookie(res.Cookies())
	}
	return nil
}
func (a *AuthSession) accept() error {
	res, err := a.http().PostJson(authBaseUrl+"repair/complete", nil)
	if err != nil {
		return fmt.Errorf("complete %s", err)
	}
	a.Auth.setHttpCookie(res.Cookies())
	return nil
}
func (a *Auth) setHttpCookie(cs []*http.Cookie) {
	_csMap := make(map[string]string)
	for _, cookie := range a.getHttpCookie() {
		_csMap[cookie.Name] = cookie.Value
	}
	for _, cookie := range cs {
		if cookie != nil && cookie.Value != "" {
			_csMap[strings.TrimSpace(cookie.Name)] = strings.TrimSpace(cookie.Value)
		}
	}
	var _csStr = ""
	for k, v := range _csMap {
		_csStr += k + "=" + v + ";"
	}
	a.Cookie = &_csStr
	return
}
func (a *Auth) getHttpCookie() (cs []*http.Cookie) {
	if a.Cookie == nil {
		return nil
	}
	for _, _c := range strings.Split(string(*a.Cookie), ";") {
		if _csArr := strings.SplitN(_c, "=", 2); len(_csArr) == 2 {
			cs = append(cs, &http.Cookie{Name: _csArr[0], Value: _csArr[1]})
		}
	}
	return
}

func (a *Auth) http() *httpclient.HttpClient {
	var client = authClient
	if a.proxy != "" {
		client = client.WithOption(httpclient.OPT_PROXY, a.proxy)
	}
	if a.AuthIP != "" {
		client = client.WithOption(httpclient.OPT_SELECT_IP, a.AuthIP)
	}
	if a.Cookie != nil {
		client = client.WithCookie(a.getHttpCookie()...)
	}
	return client
}

func (a *AuthSession) http() *httpclient.HttpClient {
	var client = a.Auth.http()
	if a.Header != nil {
		client.WithHeaders(a.Header)
	}
	return client
}
func (a *AuthSession) extractHeader(res *httpclient.Response, h ...string) {
	var hs = append([]string{"X-Apple-ID-Session-Id", "scnt"}, h...)
	if a.Header == nil {
		a.Header = make(map[string]string)
	}
	for _, s := range hs {
		a.Header[s] = res.Header.Get(s)
	}
}
func (a *AuthSession) extractMobile() error {
	res, err := a.http().Get(authBaseUrl)
	if err != nil {
		return err
	}
	for _, m := range res.ToJson("trustedPhoneNumbers").Array() {
		a.Mobiles = append(a.Mobiles, &authMobile{ID: int(m.Get("id").Int()), Mode: m.Get("pushMode").String(), Mobile: m.Get("numberWithDialCode").String()})
	}
	if res.StatusCode == 200 || res.StatusCode == 201 {
		a.SelectMobile = &authMobile{
			ID:     a.Mobiles[0].ID,
			Mode:   res.ToJson("mode").String(),
			Mobile: a.Mobiles[0].Mobile,
		}
	}
	return nil
}
