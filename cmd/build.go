package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/stubborn-gaga-0805/aurora/consts"
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
	flagOutput = flag{"output", "o", "./bin/server", "custom binaries output directory... "}
)

func newBuildCmd() *buildCmd {
	build := &buildCmd{
		baseCmd:    newBaseCmd(),
		buildFlags: new(buildFlags),
	}
	build.cmd = &cobra.Command{
		Use:     "build",
		Aliases: []string{},
		Short:   "Building your project",
		Long:    `💡 To compile/build the project, execute in the root directory of the project. eg: aurora init`,
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
			fmt.Printf("🚫 failed to create output directory [%s]...\n", outputDir)
			os.Exit(1)
			return
		}
		goArgs = append(goArgs, "-o", build.flagOutput)
	}
	if len(args) > 0 {
		absPath, err := build.FilePathToAbs(args[0])
		if err != nil {
			fmt.Printf("🚫 Command [%s] execution failed...[%v]\n", build.cmd.Use, err)
			os.Exit(1)
			return
		}
		if len(absPath) == 0 {
			fmt.Printf("🚫 The specified 'main file' [%s] does not exist...\n", args[0])
			os.Exit(1)
			return
		}
		goArgs = append(goArgs, absPath)
	} else {
		// 如果没有指定main的位置就检查是否在项目目录下
		if !build.InProjectPath() {
			fmt.Println("🚫 The 'main.go' file is not found in the current directory, please run it in the project root directory...")
			os.Exit(1)
			return
		}
		goArgs = append(goArgs, build.mainPath)
	}
	goArgs = append(goArgs, build.mainPath)
	fd := exec.Command("go", goArgs...)
	fd.Stdout = os.Stdout
	fd.Stderr = os.Stderr
	if err := fd.Run(); err != nil {
		fmt.Printf("🚫[Command: %s] execution failed...[%v]\n", build.cmd.Use, err)
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
