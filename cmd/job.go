package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/stubborn-gaga-0805/aurora/consts"
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
	flagParams   = flag{"params", "p", "", "Parameters to run the command, multiple parameters are separated by \",\"... "}
	flagShowList = flag{"list", "l", false, "View executable customer tasks"}
)

func newJobCmd() *jobCmd {
	jc := &jobCmd{
		baseCmd:  newBaseCmd(),
		jobFlags: new(jobFlags),
	}
	jc.cmd = &cobra.Command{
		Use:     "job",
		Aliases: []string{"jobs", "J", "Job", "JOB"},
		Short:   "Customer Task Related Commands",
		Long:    `💡 Customer Task Related Commands, eg: aurora job myJob -p "first_param,second_param,third_param"`,
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
		fmt.Println("🚫 The 'main.go' file is not found in the current directory, please run it in the project root directory...")
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
			fmt.Println("🚫 Please enter a task to perform...")
			os.Exit(1)
			return
		}
		goArgs = append(goArgs, "job", "-n", args[0], "-p", jc.flagParams)
	}
	fd := exec.Command(bin, goArgs...)
	fd.Stdout = os.Stdout
	fd.Stderr = os.Stderr
	if err := fd.Run(); err != nil {
		fmt.Printf("🚫[Command: %s] execution failed...[%v]\n", jc.cmd.Use, err)
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
