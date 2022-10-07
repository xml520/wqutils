package appleTools

import (
	"fmt"
	"testing"
)

func TestNewIpaUploader(t *testing.T) {
	//uploaderClient.Do()
	var header = make(map[string]string)
	header["X-Session-Id"] = "COGrDBIQpb6qNeNLSgiidvaBYPSHWw=="
	data := `{
	"storefrontId": "143465-30,28",
	"udid": "00008101-001119A42E1102AE",
	"deviceName": "Ios",
    "compressed":false
}`
	header["X-Request-Id"] = "CF72FE6F-D2C3-46DF-BF78-01DA451883E1"
	//s := "CPHEsLwlEgQIChAAEgQIDBAAEgQIBBAAEgQIERAAEgQIAhAAEgQICRAAEgQIEhAAEgQIBxAAEgQIEBAAEgQIARAAEgQICxAAEgQIAxAAEgQIBRAAEgQIBhAAEgQIDxAAEgQICBAAEgQIDRAAEgQIDhAA"
	url := "https://testflight.apple.com/v2/accounts/a83179c1-a9b1-495c-ab7c-d8a0b9aab671/apps/6443549069/builds/97536672/install"
	header["X-Active-Devices"] = ""
	//h := md5.New()
	//io.WriteString(h, header["X-Session-Id"])
	//io.WriteString(h, header["X-Request-Id"])
	//fmt.Println(hex.EncodeToString(h.Sum(nil)))
	header["User-Agent"] = "Oasis/2.6.0 OasisBuild/57 iOS/15.6.1 model/iPhone13,2 hwp/t8101 build/19G82 (6; dt:229)"
	header["X-Session-Digest"] = "b082d5b09c46e0326cf4758fe134d5bb44bd8d6a"
	res, err := uploaderClient.WithHeaders(header).Json("POST", url, data)
	if err != nil {
		fmt.Println(err, "失败", res.ToString(), res.StatusCode)
	}
	fmt.Println(header)
}
