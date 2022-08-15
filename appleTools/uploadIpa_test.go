package appleTools

import (
	"fmt"
	"testing"
)

func TestNewIpaUploader(t *testing.T) {

	newAuth := &UploadAuth{
		Account:  "1150383838@qq.com",
		Password: "yqnm-vdzw-cawl-nrkw11",
		Api: &Api{
			IssuerID: "30dc12be-550f-4da3-8d41-c775ab6b5551",
			ApiID:    "4GR84353K9",
			ApiKey:   "LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JR1RBZ0VBTUJNR0J5cUdTTTQ5QWdFR0NDcUdTTTQ5QXdFSEJIa3dkd0lCQVFRZ1h5czlpTHFPeHB1S05FQW4KK0cwM2dEWlFRdVZMeE5VYUcwNHFkckNiQy9HZ0NnWUlLb1pJemowREFRZWhSQU5DQUFTYnZQRGpWRXRuS2lhbwpCdktBSjlaeHk1cTFLQTVpUjJyNHdqbVJWbHAvejlQOVdBT05RUkNzOTNQN0M5NElVYUl4NisydWZ4YlJwYTF0CldLNE5GWFRYCi0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0=",
		},
	}
	//oldAuth := &UploadAuth{
	//	Account:  "xirangyuedujl@163.com",
	//	Password: "cigj-xuwv-rgeh-wwna",
	//}
	upload := NewIpaUploader(newAuth)
	//upload.Upload()
	err := upload.Upload("1600369627", "../ipaTools/test.ipa")
	if err != nil {
		fmt.Println("上传失败", err)
	} else {
		fmt.Println("上传成功", err)
	}
}
