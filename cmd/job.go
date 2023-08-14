package cmd

import (
	"fmt"
	"github.com/aurora/consts"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

type jobCmd struct {
	*baseCmd
	*jobFlags
}

type jobFlags struct {
	flagParams   string
	flagShowList bool
}

var (
	flagParams   = flag{"params", "p", "", "运行命令的参数, 多个参数用\",\"隔开... "}
	flagShowList = flag{"list", "l", false, "查看可执行的用户任务"}
)

func newJobCmd() *jobCmd {
	jc := &jobCmd{
		baseCmd:  newBaseCmd(),
		jobFlags: new(jobFlags),
	}
	jc.cmd = &cobra.Command{
		Use:     "job",
		Aliases: []string{"jobs", "J", "Job", "JOB"},
		Short:   "用户任务相关命令",
		Long:    `用户任务相关命令, 例如: aurora job myJob -p " 111,222"`,
		Run: func(cmd *cobra.Command, args []string) {
			jc.initJobRuntime(cmd)
			jc.run(args)
		},
	}
	addJobRuntimeFlag(jc.cmd, true)

	return jc
}

func (jc *jobCmd) initJobRuntime(cmd *cobra.Command) {
	// 检查是否在项目目录下
	if !jc.InProjectPath() {
		fmt.Println("🚫 当前目录下没有找到main文件，请在项目根目录运行...")
		os.Exit(1)
		return
	}
	jc.id, _ = os.Hostname()
	jc.env = Env(os.Getenv(consts.OSEnvKey))
	jc.jobFlags = &jobFlags{
		flagParams:   getParams(cmd),
		flagShowList: getShowList(cmd),
	}
	return
}

func (jc *jobCmd) run(args []string) {
	bin := jc.GetBin()
	goArgs := make([]string, 0)
	if jc.flagShowList {
		goArgs = append(goArgs, "job", "-l")
	} else {
		if len(args) == 0 {
			fmt.Println("🚫 请输入要执行的任务...")
			os.Exit(1)
			return
		}
		goArgs = append(goArgs, "job", "-n", args[0], "-p", jc.flagParams)
	}
	fd := exec.Command(bin, goArgs...)
	fd.Stdout = os.Stdout
	fd.Stderr = os.Stderr
	if err := fd.Run(); err != nil {
		fmt.Printf("🚫[命令: %s] 执行失败...[%v]\n", jc.cmd.Use, err)
		os.Exit(1)
		return
	}
	return
}

func addJobRuntimeFlag(cmd *cobra.Command, persistent bool) {
	getFlags(cmd, persistent).StringP(flagParams.name, flagParams.shortName, flagParams.defaultValue.(string), flagParams.usage)
	getFlags(cmd, persistent).BoolP(flagShowList.name, flagShowList.shortName, flagShowList.defaultValue.(bool), flagShowList.usage)
}

func getParams(cmd *cobra.Command) string {
	return cmd.Flag(flagParams.name).Value.String()
}

func getShowList(cmd *cobra.Command) bool {
	var (
		showList bool
		err      error
	)
	if showList, err = cmd.Flags().GetBool(flagShowList.name); err != nil {
		panic(err)
	}
	return showList
}
