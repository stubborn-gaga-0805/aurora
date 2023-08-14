package cmd

import (
	"github.com/samber/lo"
	"github.com/stubborn-gaga/aurora/consts"
)

type Env string

type AppName string

type Version string

type ConfigFilePath string

var (
	supportsRuntimeEnvs = []Env{consts.EnvLocal, consts.EnvDev, consts.EnvTest, consts.EnvPre, consts.EnvProd}
	debugEnvs           = []Env{consts.EnvLocal, consts.EnvDev, consts.EnvTest, consts.EnvPre, consts.EnvProd}
)

func (e Env) Check() bool {
	return lo.Contains(supportsRuntimeEnvs, e)
}

func (e Env) IsDebug() bool {
	return lo.Contains(debugEnvs, e)
}

func (e Env) ToString() string {
	return string(e)
}

func (e AppName) ToString() string {
	return string(e)
}

func (e Version) ToString() string {
	return string(e)
}

func (e ConfigFilePath) ToString() string {
	return string(e)
}

func (e ConfigFilePath) UserDefined() bool {
	return len(e.ToString()) > 0
}
