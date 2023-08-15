package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/stubborn-gaga-0805/aurora/consts"
	"github.com/stubborn-gaga-0805/aurora/helpers"
	"os"
	"os/exec"
)

type initCmd struct {
	*baseCmd
	goModPath string
}

func newInitCmd() *initCmd {
	init := &initCmd{
		baseCmd: newBaseCmd(),
	}
	init.cmd = &cobra.Command{
		Use:     "init",
		Aliases: []string{},
		Short:   "Initialize the project",
		Long:    "💡 Initialize the project, eg: aurora init",
		Run: func(cmd *cobra.Command, args []string) {
			init.initInitRuntime()
			init.run()
		},
	}

	return init
}

func (init *initCmd) initInitRuntime() {
	var err error
	// 检查是否在项目目录下
	if !init.InProjectPath() {
		fmt.Println("🚫 The 'main.go' file is not found in the current directory, please run it in the project root directory...")
		os.Exit(1)
		return
	}
	if init.goModPath, err = init.FilePathToAbs("./go.mod"); err != nil {
		fmt.Printf("🚫[Command: %s] execution failed...[%v]\n", init.cmd.Use, err)
		os.Exit(1)
		return
	}
	if len(init.goModPath) == 0 {
		fmt.Printf("🚫 The 'go.mod' file was not found in the current directory, initialization failed！")
		os.Exit(1)
		return
	}
	init.id, _ = os.Hostname()
	init.env = Env(os.Getenv(consts.OSEnvKey))
	return
}

func (init *initCmd) run() {
	var (
		success = 0
		cmdList = []*exec.Cmd{
			exec.Command("go", "install", "github.com/google/wire/cmd/wire@latest"),
			exec.Command("go", "install", "google.golang.org/protobuf/cmd/protoc-gen-go@latest"),
			exec.Command("go", "install", "google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"),
			exec.Command("go", "install", "github.com/google/gnostic/cmd/protoc-gen-openapi@latest"),
			exec.Command("go", "install", "gorm.io/gen/tools/gentool@latest"),
			exec.Command("go", "mod", "tidy"),
			exec.Command("go", "mod", "verify"),
		}
	)
	bar := helpers.NewProgressBar(len(cmdList), "Initializing project...")
	for _, cmd := range cmdList {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("\n🚫 %v %s [%v]\n", cmd.Args, color.RedString("execution failed!"), err)
			continue
		} else {
			success++
			fmt.Printf("\n✅ %v %s", cmd.Args, color.GreenString("all execution succeed!"))
			bar.Increment()
		}
	}
	if success != len(cmdList) {
		fmt.Printf("\n\n‼️ %s", color.YellowString("partial success!"))
		os.Exit(1)
		return
	}
	bar.Finish()
	fmt.Printf("\n\n🍺🍺🍺 Initialize the project successfully!\n")

	return
}
