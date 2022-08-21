package dxCaptcha

import (
	"fmt"
	"testing"
)

func TestNewCaptchaClient(t *testing.T) {
	appId := "1ac59bb3e5f81f4ad80358f4b09b9754"
	appSecret := "11a38843a6b40e3a8f4efb4bbb0119d9"
	captchaClient := NewCaptchaClient(appId, appSecret)
	captchaClient.SetCaptchaUrl("https://vip6.dingxiang-inc.com/api/tokenVerify")
	captchaResponse := captchaClient.VerifyToken("828A5D89F200E4149894FBCBAC6F1E563A47D9BB22F13966369A514720C0E44611CF5241EA391B7B52186296C9159AFAD7BF652554E6DE44052462FA15A189191A6E8F449D5FCEE7C7B9865A61B87206:62fccd68upipZk3thtCGdycuG9Ftp1ySGCRVcnW1")
	if captchaResponse.Result {
		fmt.Println("验证成功")
		/*token验证通过，继续其他流程*/
	} else {
		fmt.Println("验证失败")
		/*token验证失败，业务系统可以直接阻断该次请求或者继续弹验证码*/
	}
}
