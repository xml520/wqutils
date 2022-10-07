package appleTools

import (
	"errors"
	"fmt"
	"github.com/xml520/wqutils/httpclient"
	"howett.net/plist"
	"strings"
)

const (
	itunesAuthUrl  = `https://p25-buy.itunes.apple.com/WebObjects/MZFinance.woa/wa/authenticate?guid=38F9D3AF4E61`
	itunesAuthBody = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>appleId</key>
	<string>%s</string>
	<key>attempt</key>
	<string>4</string>
	<key>createSession</key>
	<string>true</string>
	<key>guid</key>
	<string>38F9D3AF4E60</string>
	<key>password</key>
	<string>%s</string>
	<key>rmp</key>
	<string>0</string>
	<key>why</key>
	<string>signIn</string>
</dict>
</plist>`
)

type Itunes map[string]any

func ItunesLogin(email string, password string) (Itunes, error) {
	if res, err := httpclient.Do("POST", itunesAuthUrl, map[string]string{
		"Content-Type":    "application/x-www-form-urlencoded",
		"Cookie":          "itspod=6;",
		"Accept":          "*/*",
		"User-Agent":      "Configurator/2.15 (Macintosh; OS X 11.0.0; 16G29) AppleWebKit/2603.3.8",
		"Accept-Language": "zh-CN,zh-Hans;q=0.9",
	}, strings.NewReader(fmt.Sprintf(itunesAuthBody, email, password))); err != nil {
		return nil, err
	} else {
		var s Itunes
		if _, err = plist.Unmarshal([]byte(res.ToString()), &s); err != nil {
			return nil, err
		}
		if _, ok := s["cancel-purchase-batch"]; ok {
			return nil, errors.New("登录失败")
		}
		return s, nil
	}

}
func (i Itunes) GeyMoney() (any, error) {
	if m, ok := i["creditDisplay"]; ok {
		return m.(string), nil
	} else {
		return nil, errors.New("字段不存在")
	}
}
