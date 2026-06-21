package feishu

import "errors"

type SaveConfigRequest struct {
	AppId     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}

type GetConfigResponse struct {
	AppId           string `json:"app_id"`
	AppSecretMasked string `json:"app_secret_masked"`
	IsValid         bool   `json:"is_valid"`
	ValidMsg        string `json:"valid_msg"`
	HasToken        bool   `json:"has_token"`
}

type CallbackResponse struct {
	Message string `json:"message"`
}

var (
	ErrCodeEmpty        = errors.New("缺少code参数")
	ErrRedirectURIEmpty = errors.New("缺少redirect_uri参数")
	ErrConfigEmpty      = errors.New("环境变量里缺少FEISHU_ROBOT_APP_ID或FEISHU_ROBOT_APP_SECRET, 请配置环境变量")
	ErrTokenEmpty       = errors.New("user_access_token为空")
)
