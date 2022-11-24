package appleTools

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/xml520/wqutils/httpclient"
	"log"
	"strings"
	"time"
)

var apiClient *httpclient.HttpClient

const (
	tokenExpire = 1000
	apiBaseurl  = "https://api.appstoreconnect.apple.com/v1/"
)

//func init() {
//	apiClient = httpclient.NewHttpClient().Defaults(map[interface{}]interface{}{
//		httpclient.OPT_COOKIEJAR: false,
//		"Accept":                 jsonContentType,
//		httpclient.OPT_AFTER_REQUEST_FUNC: func(res *httpclient.Response) error {
//			if res == nil {
//				return errors.New("请求错误")
//			}
//			if res.StatusCode < 299 {
//				return nil
//			}
//			if msg := res.ToJson("errors.0.detail").String(); msg != "" {
//				return errors.New(msg)
//			}
//			return errors.New(fmt.Sprintf("%s 未知错误 状态码：%v", res.Request.URL.String(), res.Status))
//		},
//		httpclient.OPT_TIMEOUT: 30,
//	})
//}

type Api struct {
	IssuerID string `json:"issuer_id" gorm:"index;comment:IssuerID"`
	ApiID    string `json:"api_id" gorm:"index;comment:ApiID"`
	ApiKey   string `json:"api_key" gorm:"type:text;comment:ApiKey"`
}

func newApiClient() *httpclient.HttpClient {
	return httpclient.NewHttpClient().Defaults(map[interface{}]interface{}{
		httpclient.OPT_COOKIEJAR: false,
		"Accept":                 jsonContentType,
		httpclient.OPT_AFTER_REQUEST_FUNC: func(res *httpclient.Response) error {
			if res == nil {
				return errors.New("请求错误")
			}
			if res.StatusCode < 299 {
				return nil
			}
			if msg := res.ToJson("errors.0.detail").String(); msg != "" {
				return errors.New(msg)
			}
			return errors.New(fmt.Sprintf("%s 未知错误 状态码：%v", res.Request.URL.String(), res.Status))
		},
		httpclient.OPT_TIMEOUT: 30,
	})
}

func (a *Api) Do(method, url string, data any) (*httpclient.Response, error) {
	token, err := a.generateToken(tokenExpire)
	if err != nil {
		return nil, err
	}
	if strings.ToTitle(method) == "GET" {
		data = ""
	}
	return newApiClient().WithHeader("Authorization", "Bearer "+token).Json(method, apiBaseurl+url, data)
}

func (a *Api) http() *httpclient.HttpClient {
	token, err := a.generateToken(tokenExpire)
	if err != nil {
		log.Println("token 生成失败", err)
	}
	return newApiClient().WithHeader("Authorization", "Bearer "+token)
}
func (a *Api) generateToken(expire int64) (string, error) {
	expires := time.Now().Unix() + expire // 19分钟有效期
	token := jwt.Token{
		Header: map[string]interface{}{
			"typ": "JWT",
			"alg": "ES256",
			"kid": a.ApiID,
		},
		Claims: jwt.StandardClaims{
			Audience:  "appstoreconnect-v1",
			Issuer:    a.IssuerID,
			ExpiresAt: expires,
		},
		Method: jwt.SigningMethodES256,
	}
	var pk *ecdsa.PrivateKey
	if rawByte, err := base64.StdEncoding.DecodeString(a.ApiKey); err != nil {
		return "", err
	} else {
		if block, _ := pem.Decode(rawByte); block == nil {
			return "", errors.New("密钥文件不是pem格式")
		} else {
			if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err != nil {
				return "", err
			} else {
				switch key.(type) {
				case *ecdsa.PrivateKey:
					pk = key.(*ecdsa.PrivateKey)
				default:
					return "", errors.New("AuthKey的类型必须为ecdsa.PrivateKey")
				}
			}
		}

	}
	return token.SignedString(pk)
}
