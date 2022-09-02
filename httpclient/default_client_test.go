// Copyright 2014-2019 Liu Dong <ddliuhb@gmail.com>.
// Licensed under the MIT license.

package httpclient

import (
	"fmt"
	"net/http"
	"testing"
)

func TestDefaultClient(t *testing.T) {
	newC := NewHttpClient().Defaults(map[interface{}]interface{}{
		"Content-Type": "application/json",
	})
	res, err := newC.WithHeader("cookie", "v=1").WithCookie(&http.Cookie{Name: "t", Value: "121212"}).Post("http://127.0.0.1:12345/api/ping", nil)
	if err != nil {
		fmt.Println("请求失败")
	}
	fmt.Println("请求头", res.Request.Header)
	res, err = newC.Post("http://127.0.0.1:12345/api/ping", nil)
	fmt.Println("请求头2", res.Request.Header)
}
func ss(s []string) []string {
	fmt.Println("ok")
	return s
}
