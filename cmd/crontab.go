package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/stubborn-gaga-0805/aurora/consts"
	"os"
	"os/exec"
)

type cronCmd struct {
	*baseCmd
	*crontabFlags
}

type crontabFlags struct {
	crontabList bool
}

var (
	flagCrontabList = flag{"list", "l", false, "List running crontab tasks"}
)

func newCronCmd() *cronCmd {
	c := &cronCmd{newBaseCmd(), new(crontabFlags)}
	c.cmd = &cobra.Command{
		Use:     "cron",
		Aliases: []string{"crontab", "Cron", "Crontab", "CRON"},
		Short:   "Crontab task related commands",
		Long:    "",
		Run: func(cmd *cobra.Command, args []string) {
			c.initCrontabRuntime(cmd)
			c.initConfig()
			c.run()
		},
	}
	addCrontabRuntimeFlag(c.cmd, true)

	return c
}

func (c *cronCmd) initCrontabRuntime(cmd *cobra.Command) {
	// 检查是否在项目目录下
	if !c.InProjectPath() {
		fmt.Println("🚫 The 'main.go' file is not found in the current directory, please run it in the project root directory...")
		os.Exit(1)
		return
	}

	c.id, _ = os.Hostname()
	c.env = Env(os.Getenv(consts.OSEnvKey))
	c.configFilePath = fmt.Sprintf("./configs/config.%s.yaml", c.env)
	c.crontabFlags = &crontabFlags{
		crontabList: getCrontabList(c.cmd),
	}
	return
}

func (c *cronCmd) run() {
	var (
		bin    = c.GetBin()
		goArgs = []string{"cron"}
	)
	if c.crontabList {
		goArgs = append(goArgs, "-l")
	}
	fd := exec.Command(bin, goArgs...)
	fd.Stdout = os.Stdout
	fd.Stderr = os.Stderr
	if err := fd.Run(); err != nil {
		fmt.Printf("🚫Command [%s] execution failed...[%v]\n", c.cmd.Use, err)
		os.Exit(1)
		return
	}
	return
}

func addCrontabRuntimeFlag(cmd *cobra.Command, persistent bool) {
	getFlags(cmd, persistent).BoolP(flagCrontabList.name, flagCrontabList.shortName, flagCrontabList.defaultValue.(bool), flagCrontabList.usage)
}

func getCrontabList(cmd *cobra.Command) bool {
	var (
		crontabList bool
		err         error
	)
	if crontabList, err = cmd.Flags().GetBool(flagCrontabList.name); err != nil {
		panic(err)
	}
	return crontabList
}
