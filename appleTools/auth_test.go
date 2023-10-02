package appleTools

import (
	"fmt"
	"github.com/dop251/goja"
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

func TestAuth_SignInV2(t *testing.T) {
	const script = `
    function fib() {
       return 5555555555555555555555555555555555555555555555555555555555555555555555555555 >> 1
    }
    `
	vm := goja.New()
	_, err := vm.RunString(script)
	if err != nil {
		fmt.Println("JS代码有问题！")
		return
	}
	var fn func(int32) int32
	err = vm.ExportTo(vm.Get("fib"), &fn)
	if err != nil {
		fmt.Println("Js函数映射到 Go 函数失败！")
		return
	}
	fmt.Println("斐波那契数列第30项的值为：", fn(30))
	fmt.Println(8 >> 1)
	return
	a := Auth{
		Account:  "1150383838@qq.com",
		Password: "QWas649896461..",
	}
	s, err := a.SignInV2()
	if err != nil {
		fmt.Println("登录失败", err)
		fmt.Printf("%+v %+v", s, s.Mobiles)
	} else {
		fmt.Printf("%+v %+v", s, s.Mobiles)
	}
}
