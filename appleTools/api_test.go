package appleTools

import (
	"fmt"
	"testing"
)

func TestApi(t *testing.T) {
	api := &Api{
		IssuerID: "49d4df47-dba3-4777-90f4-5822f5994dd7",
		ApiID:    "M3UNDRJSS3",
		ApiKey:   "LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JR1RBZ0VBTUJNR0J5cUdTTTQ5QWdFR0NDcUdTTTQ5QXdFSEJIa3dkd0lCQVFRZ3VnekFtV25PNXR4VDFGei8KSUNLRkxCdnloVmhyakgwNnE4eldGMUR1dTZLZ0NnWUlLb1pJemowREFRZWhSQU5DQUFSbVc4OVFjNUVMZmlCOQoyZGU1TkNlVmhSMExGNUdWa0p3WXV6emdoNGdhekI5VHN1V3dyN2U1ZjZCdU5yb3JMcEZtbE91d3hVVkR2dlR2CmhkN1Q1bjQwCi0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0=",
	}
	res, err := api.Do("POST", "userInvitations", map[string]any{
		"data": map[string]any{
			"type": "userInvitations",
			"attributes": map[string]any{
				"allAppsVisible": true,
				"email":          "3377777722@qq.com",
				"firstName":      "ID1ID",
				"lastName":       "SSS",
				"roles":          []string{"ADMIN"},
			},
		},
	})
	if err != nil {
		fmt.Println("请求失败", err)
	} else {
		fmt.Println("请求成功", res.ToString())

	}
}
