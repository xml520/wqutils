// Copyright 2014-2019 Liu Dong <ddliuhb@gmail.com>.
// Licensed under the MIT license.

package httpclient

import (
	"fmt"
	"net/url"
	"testing"
)

func TestDefaultClient(t *testing.T) {
	for i, s := range ss([]string{"SS", "ssss"}) {
		fmt.Println(i, s)
	}
	d := NewHttpClient()
	res, err := d.Get("http://lumtest.com/myip.json")

	if err != nil {
		t.Error("get failed", err)
	}

	if res.StatusCode != 200 {
		t.Error("Status Code not 200")
	}
	u, _ := url.Parse("http://127.0.0.1")
	u.User = url.UserPassword("121212112", "1212121")

	fmt.Println(res.ToString(), u.String())
}
func ss(s []string) []string {
	fmt.Println("ok")
	return s
}
