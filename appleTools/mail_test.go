package appleTools

import (
	"fmt"
	"testing"
)

type tt struct {
}

func (t tt) VerifyAppleID(idType *EmailVerifyAppleIDType) {
	//TODO implement me
	//panic("implement me")
}

func (t tt) TestflightLink(linkType *EmailTestflightLinkType) {
	//TODO implement me
}

func (t tt) BuildFailed(failedType *EmailBuildFailedType) {
	//TODO implement me
}

func (t tt) VerifyEmail(emailType *EmailVerifyEmailType) {
	//TODO implement me

}

func (t tt) NoType(noType *EmailNoType) {
	//TODO implement me

	fmt.Println("没有识别", noType.Subject)
}

func (t tt) TeamInvite(inviteType *EmailTeamInviteType) {
	//TODO implement me

}

func TestName(t *testing.T) {
	//MailServerListen(":25", new(tt))
	//s := "c82cf33044f447ba81f52adc90722f9437a8b2ae42274e7884aa22e10752b59f61472c39"
	//fmt.Println(strconv.ParseUint(s[len(s)-8:], 16, 0))
	//s, _ := regexp.Compile("[0-9]{6}")
	//log.Println(s.FindString(`无法识别 huanyu109@testflight.li 验证您的 Apple ID 电子邮件地址 您已选                  择此电子邮件地址作为您的新 Apple ID。为验证此电子邮件地址属于您，请在电子邮件验                                    证页面输入下方验证码：

	//MailServerListen(":25",ss)
	//ss.BuildFailed(&EmailBuildFailedType{})

}
