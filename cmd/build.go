package cmd

import (
	"fmt"
	"github.com/aurora/consts"
	"github.com/spf13/cobra"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

type buildCmd struct {
	*baseCmd
	*buildFlags
}

type buildFlags struct {
	flagOutput string
}

var (
	flagOutput = flag{"output", "o", "./bin/server", "自定义二进制文件输出... "}
)

func newBuildCmd() *buildCmd {
	build := &buildCmd{
		baseCmd:    newBaseCmd(),
		buildFlags: new(buildFlags),
	}
	build.cmd = &cobra.Command{
		Use:     "build",
		Aliases: []string{},
		Short:   "编译项目",
		Long:    "编译项目, 进入项目根目录后: aurora init",
		Run: func(cmd *cobra.Command, args []string) {
			build.initBuildRuntime(cmd)
			build.run(args)
		},
	}
	addBuildRuntimeFlag(build.cmd, true)

	return build
}

func (build *buildCmd) initBuildRuntime(cmd *cobra.Command) {
	build.id, _ = os.Hostname()
	build.env = Env(os.Getenv(consts.OSEnvKey))
	build.buildFlags = &buildFlags{
		flagOutput: getOutput(cmd),
	}
	return
}

func (build *buildCmd) run(args []string) {
	var (
		goArgs    = []string{"build", "-ldflags=-s -w"}
		outputDir = ""
	)
	if len(build.flagOutput) > 0 {
		outputDir = filepath.Dir(build.flagOutput)
		if err := os.MkdirAll(outputDir, fs.ModePerm); err != nil {
			fmt.Printf("🚫 创建输出目录[%s]失败...\n", outputDir)
			os.Exit(1)
			return
		}
		goArgs = append(goArgs, "-o", build.flagOutput)
	}
	if len(args) > 0 {
		absPath, err := build.FilePathToAbs(args[0])
		if err != nil {
			fmt.Printf("🚫[命令: %s] 执行失败...[%v]\n", build.cmd.Use, err)
			os.Exit(1)
			return
		}
		if len(absPath) == 0 {
			fmt.Printf("🚫 指定的main文件[%s]不存在...\n", args[0])
			os.Exit(1)
			return
		}
		goArgs = append(goArgs, absPath)
	} else {
		// 如果没有指定main的位置就检查是否在项目目录下
		if !build.InProjectPath() {
			fmt.Println("🚫 当前目录下没有找到main文件，请在项目根目录运行...")
			os.Exit(1)
			return
		}
		goArgs = append(goArgs, build.mainPath)
	}
	fd := exec.Command("go", goArgs...)
	fd.Stdout = os.Stdout
	fd.Stderr = os.Stderr
	if err := fd.Run(); err != nil {
		fmt.Printf("🚫[命令: %s] 执行失败...[%v]\n", build.cmd.Use, err)
		os.Exit(1)
		return
	}
	return
}

func addBuildRuntimeFlag(cmd *cobra.Command, persistent bool) {
	getFlags(cmd, persistent).StringP(flagOutput.name, flagOutput.shortName, flagOutput.defaultValue.(string), flagOutput.usage)
}

func getOutput(cmd *cobra.Command) string {
	return cmd.Flag(flagOutput.name).Value.String()
}
