package dxCaptcha

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type CaptchaClient struct {
	captchaUrl string
	appId      string
	appSecret  string
	timeout    int
}

// SetTimeout 设置超时
func (c *CaptchaClient) SetTimeout(timeout int) {
	c.timeout = timeout
}

// SetCaptchaUrl 设置api地址
func (c *CaptchaClient) SetCaptchaUrl(captchaUrl string) {
	c.captchaUrl = captchaUrl
}

// NewCaptchaClient 初始化验证码
func NewCaptchaClient(appId, appSecret string) CaptchaClient {
	return CaptchaClient{"https://cap.dingxiang-inc.com/api/tokenVerify", appId, appSecret, 2000}
}

func NewCaptchaClientByRequestConfig(appId, appSecret string, timeout int) CaptchaClient {
	return CaptchaClient{"https://cap.dingxiang-inc.com/api/tokenVerify", appId, appSecret, timeout}
}

// 无感验证响应接口
type verifyTokenResponse struct {
	Success bool
	Msg     string
	Ip      string
	Code    string
	Tpc     string
	Uid     string
}

func isBlank(s string) bool {
	if len(s) != 0 && len(strings.TrimSpace(s)) != 0 {
		return false
	}
	return true
}
func (c *CaptchaClient) VerifyToken(token string) CaptchaResponse {
	return c.getResponse(token, "")
}
func (c *CaptchaClient) VerifyTokenAndIP(token, ip string) CaptchaResponse {
	return c.getResponse(token, ip)
}

func (c *CaptchaClient) getResponse(token, ip string) CaptchaResponse {
	if isBlank(token) || isBlank(c.appId) || isBlank(c.appSecret) || len(token) > 1024 {
		return CaptchaResponse{Result: false, CaptchaStatus: string(WRONG_PARAMETER)}
	}
	args := strings.Split(token, ":")
	// 获取 MD5 加密结果
	sum := md5.Sum([]byte(c.appSecret + args[0] + c.appSecret))
	sign := hex.EncodeToString(sum[:])
	// 获取AppKey
	var key string
	if len(args) == 2 {
		key = args[1]
	}

	reqUrl := fmt.Sprintf("%s?token=%s&constId=%s&appKey=%s&sign=%s&ip=%s", c.captchaUrl, args[0], key, c.appId, sign, ip)
	client := http.Client{Timeout: time.Duration(c.timeout) * time.Millisecond}
	return getHttpResponse(reqUrl, client)
}

func getHttpResponse(reqUrl string, client http.Client) CaptchaResponse {
	resp, err := client.Get(reqUrl)
	if err != nil {
		return CaptchaResponse{Result: true, CaptchaStatus: "server connect error:" + fmt.Sprint(err)}
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		// IO错误
		if err != nil {
			return CaptchaResponse{Result: true, CaptchaStatus: "server connect error:" + fmt.Sprint(err)}
		}
		// 风控引擎响应结果
		var response verifyTokenResponse
		// JSON转换错误
		if json.Unmarshal(body, &response) != nil {
			return CaptchaResponse{Result: true, CaptchaStatus: "server connect error:" + fmt.Sprint(err)}
		}
		return CaptchaResponse{
			Result:        response.Success,
			CaptchaStatus: string(SERVER_SUCCESS),
			Ip:            response.Ip,
			Tpc:           response.Tpc,
			Uid:           response.Uid,
			Code:          response.Code,
		}
	} else {
		return CaptchaResponse{Result: true, CaptchaStatus: string(SERVER_FAILED)}
	}
}
