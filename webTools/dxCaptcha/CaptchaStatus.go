package dxCaptcha

// CaptchaStatus 验证码响应状态枚举
type CaptchaStatus string

const (
	SERVER_SUCCESS  CaptchaStatus = "SERVER_SUCCESS"
	SERVER_FAILED   CaptchaStatus = "SERVER_FAILED"
	WRONG_PARAMETER CaptchaStatus = "WRONG_PARAMETER"
)
