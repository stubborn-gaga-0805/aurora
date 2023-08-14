package cmd

import (
	"context"
	"fmt"
	"github.com/aurora/conf"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"path/filepath"
)

type baseCmd struct {
	cmd *cobra.Command
	ctx context.Context

	id             string
	isDebug        bool
	env            Env
	configFilePath string
	workingDir     string
	mainPath       string
	binPath        string
	hasBin         bool
}

func newBaseCmd() *baseCmd {
	var (
		err   error
		id, _ = os.Hostname()
		bc    = &baseCmd{
			id:     id,
			ctx:    context.Background(),
			hasBin: true,
		}
	)
	bc.workingDir, err = os.Getwd()
	if err != nil {
		fmt.Printf("🚧 Stopped...[%v]\n", err)
		os.Exit(1)
		return nil
	}
	bc.mainPath = filepath.Join(bc.workingDir, "main.go")
	// 检查bin文件
	bc.binPath = filepath.Join(bc.workingDir, "/bin/server")
	_, err = os.Stat(bc.binPath)
	if os.IsNotExist(err) {
		bc.hasBin = false
	}
	return bc
}

func (base *baseCmd) getCmd() *cobra.Command {
	return base.cmd
}

func (base *baseCmd) addCommands(commands ...cmder) {
	for _, command := range commands {
		base.cmd.AddCommand(command.getCmd())
	}
}

func (base *baseCmd) initConfig() {
	var (
		configs *conf.App
		err     error
	)

	// 设置配置文件
	viper.SetConfigType("yaml")
	viper.SetConfigFile(base.configFilePath)
	// 读取配置文件到结构体
	if err = viper.ReadInConfig(); err != nil {
		panic(err)
	}
	if err = viper.Unmarshal(&configs); err != nil {
		panic(err)
	}
	conf.SetConfig(configs)

	return
}

func (base *baseCmd) GetBin() string {
	if base.hasBin {
		return base.binPath
	}
	return base.Build()
}

func (base *baseCmd) Build() string {
	fd := exec.Command("go", "build", "ldflags=\"-s -w\"", "-o", base.binPath, base.mainPath)
	fd.Stdout = os.Stdout
	fd.Stderr = os.Stderr
	if err := fd.Run(); err != nil {
		fmt.Printf("🚫 Build失败...[%v]\n", err)
		os.Exit(1)
		return ""
	}
	base.hasBin = true
	return base.binPath
}

func (base *baseCmd) InProjectPath() bool {
	_, err := os.Stat(base.mainPath)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func (base *baseCmd) FilePathToAbs(path string) (absPath string, err error) {
	if absPath, err = filepath.Abs(path); err != nil {
		return "", err
	}
	if _, err = os.Stat(absPath); os.IsNotExist(err) {
		return "", nil
	}
	return absPath, nil
}

type flag struct {
	name         string
	shortName    string
	defaultValue interface{}
	usage        string
}

func getFlags(cmd *cobra.Command, persistent bool) *pflag.FlagSet {
	flags := cmd.Flags()
	if persistent {
		flags = cmd.PersistentFlags()
	}
	return flags
}
