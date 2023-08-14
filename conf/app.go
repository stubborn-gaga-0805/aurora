package conf

type App struct {
	Env Env `json:"env"`
}

var (
	runtimeConfig  *App
	configFilePath string
)

func SetConfig(cfg *App) {
	runtimeConfig = cfg
}

func GetConfig() *App {
	return runtimeConfig
}

func SetConfigPath(path string) {
	configFilePath = path
}

func GetConfigPath() string {
	return configFilePath
}
