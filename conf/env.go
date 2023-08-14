package conf

type Env struct {
	AppId      string `json:"appId"`
	AppName    string `json:"appName"`
	AppVersion string `json:"appVersion"`
	AppEnv     string `json:"appEnv"`
}
