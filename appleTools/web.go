package appleTools

import (
	"errors"
	"fmt"
	"github.com/xml520/wqutils/httpclient"
	"strings"
)

type Web struct {
	Cookie string `json:"cookie" gorm:"type:text;comment:Cookie"`
	AuthIP string `json:"auth_ip" gorm:"comment:登录IP"` // 选择ip
}

func newWebClient() *httpclient.HttpClient {
	return httpclient.NewHttpClient().Defaults(map[interface{}]interface{}{
		"Accept-Language":        language,
		"Accept":                 jsonContentType,
		httpclient.OPT_COOKIEJAR: false,
		httpclient.OPT_AFTER_REQUEST_FUNC: func(res *httpclient.Response) error {
			if res == nil {
				return errors.New("请求错误")
			}
			if res.StatusCode < 299 {
				return nil
			}
			switch res.StatusCode {
			case 401:
				return errors.New("cookie已过期")
			default:
				if msg := res.ToJson("errors.0.detail").String(); msg != "" {
					return errors.New(msg)
				}
				return errors.New(fmt.Sprintf("%s 未知错误 状态码：%v", res.Request.URL.String(), res.Status))
			}
		},
		httpclient.OPT_TIMEOUT: 30,
	})
}
func (w *Web) http() *httpclient.HttpClient {
	var client = newWebClient()
	if w.AuthIP != "" {

		if strings.Index(w.AuthIP, "http") != -1 {
			//fmt.Println(w.AuthIP)
			client.WithOption(httpclient.OPT_PROXY, w.AuthIP)
		} else {
			client.WithOption(httpclient.OPT_SELECT_IP, w.AuthIP)
		}
	}
	if w.Cookie != "" {
		client.WithHeader("cookie", w.Cookie)
	}
	return client
}
func (w *Web) Http() *httpclient.HttpClient {
	return w.http()
}
func (w *Web) SetWebIP(ip string) *Web {
	w.AuthIP = ip
	return w
}
func (w *Web) Do(method string, url string, data any) (*httpclient.Response, error) {
	if strings.ToTitle(method) == "GET" {
		data = ""
	}
	return w.http().Json(method, url, data)
}
