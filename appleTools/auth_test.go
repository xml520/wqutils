package appleTools

import (
	"fmt"
	"testing"
)

func TestAuth_SignIn(t *testing.T) {
	a := Auth{
		Account:  "1150383838@qq.com",
		Password: "QWas649896461..",
	}
	s, err := a.SignIn()
	if err != nil {
		fmt.Println("登录失败", err)
		fmt.Printf("%+v %+v", s, s.Mobiles)
	} else {
		fmt.Printf("%+v %+v", s, s.Mobiles)
	}
}
