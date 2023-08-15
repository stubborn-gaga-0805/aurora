package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

type runCmd struct {
	*baseCmd
	*runFlags
}

type runFlags struct {
	appName        string
	appEnv         string
	appVersion     string
	withCronJob    bool
	withWs         bool
	withEtcdConfig bool
	withoutHttp    bool
	withoutMQ      bool
}

var (
	flagAppName        = flag{"name", "n", "prepare-to-go", "Set application name"}
	flagAppEnvironment = flag{"env", "e", "local", "Set the operating environment of the application"}
	flagAppVersion     = flag{"version", "v", "v1.0", "Set the version of the application"}
	flagAppConfig      = flag{"config", "c", "", "Set the path to the configuration file"}
	flagWithCronJob    = flag{"with.cron", "", false, "Whether to start the crontab task"}
	flagWithWs         = flag{"with.ws", "", false, "Whether to start the websocket server"}
	flagWithoutHttp    = flag{"without.server", "", false, "Do not start the http server"}
	flagWithoutMQ      = flag{"without.mq", "", false, "Do not start the MQ server"}
)

func newRunCmd() *runCmd {
	run := &runCmd{newBaseCmd(), new(runFlags)}
	run.cmd = &cobra.Command{
		Use:     "run",
		Aliases: []string{"start", "running", "up"},
		Short:   "Start web server (such as: http, grpc, websocket), and start Http server by default",
		Long:    `💡 Start your app... eg: aurora run -n myApp e test --with.cron --without.mq`,
		Run: func(cmd *cobra.Command, args []string) {
			run.initRuntime(cmd)
			run.initConfig()
			run.run()
		},
	}
	addServerRuntimeFlag(run.cmd, true)

	return run
}

// 初始化运行时环境
func (run *runCmd) initRuntime(cmd *cobra.Command) {
	// 检查是否在项目目录下
	if !run.InProjectPath() {
		fmt.Println("🚫 The 'main.go' file is not found in the current directory, please run it in the project root directory...")
		os.Exit(1)
		return
	}
	var (
		err      error
		appEnv   = getAppEnvironment(cmd)
		runFlags = &runFlags{
			appName:    string(getAppName(cmd)),
			appVersion: string(getAppVersion(cmd)),
		}
	)
	run.id, _ = os.Hostname()
	run.env = appEnv
	if !appEnv.Check() {
		panic(fmt.Sprintf("Unsupported operating environment... 【%s】", run.env))
	}
	run.isDebug = appEnv.IsDebug()
	run.runFlags = runFlags
	run.runFlags.appEnv = run.env.ToString()

	// 初始化配置信息
	configPath := getAppConfigPath(cmd)
	if configPath.UserDefined() {
		run.configFilePath = configPath.ToString()
	} else {
		run.configFilePath = fmt.Sprintf("./configs/config.%s.yaml", run.env)
	}
	run.runFlags.withCronJob = getWithCronJob(run.cmd)
	run.runFlags.withWs = getWithWs(run.cmd)
	run.runFlags.withoutHttp = getWithOutHttp(run.cmd)
	run.runFlags.withoutMQ = getWithoutMQConfig(run.cmd)
	if err != nil {
		panic(err)
	}

	return
}

func (run *runCmd) run() {
	bin := run.Build()
	goArgs := []string{
		"run",
		"-c", run.configFilePath,
		"-e", run.runFlags.appEnv,
	}
	if run.runFlags.withWs {
		fmt.Printf("--%s\n", flagWithWs.name)
		goArgs = append(goArgs, fmt.Sprintf("--%s", flagWithWs.name))
	}
	if run.runFlags.withCronJob {
		fmt.Printf("--%s\n", flagWithCronJob.name)
		goArgs = append(goArgs, fmt.Sprintf("--%s", flagWithCronJob.name))
	}
	if run.runFlags.withoutHttp {
		fmt.Printf("--%s\n", flagWithoutHttp.name)
		goArgs = append(goArgs, fmt.Sprintf("--%s", flagWithoutHttp.name))
	}
	if run.runFlags.withoutMQ {
		fmt.Printf("--%s\n", flagWithoutMQ.name)
		goArgs = append(goArgs, fmt.Sprintf("--%s", flagWithoutMQ.name))
	}
	fd := exec.Command(bin, goArgs...)
	fd.Stdout = os.Stdout
	fd.Stderr = os.Stderr
	if err := fd.Run(); err != nil {
		fmt.Printf("🚫[服务: %s] 启动失败...[err: %v]\n", run.appName, err)
		os.Exit(1)
		return
	}
	return
}

// 通过命令注入运行环境参数
func addServerRuntimeFlag(cmd *cobra.Command, persistent bool) {
	getFlags(cmd, persistent).StringP(flagAppName.name, flagAppName.shortName, flagAppName.defaultValue.(string), flagAppName.usage)
	getFlags(cmd, persistent).StringP(flagAppEnvironment.name, flagAppEnvironment.shortName, flagAppEnvironment.defaultValue.(string), flagAppEnvironment.usage)
	getFlags(cmd, persistent).StringP(flagAppVersion.name, flagAppVersion.shortName, flagAppVersion.defaultValue.(string), flagAppVersion.usage)
	getFlags(cmd, persistent).StringP(flagAppConfig.name, flagAppConfig.shortName, flagAppConfig.defaultValue.(string), flagAppConfig.usage)
	getFlags(cmd, persistent).BoolP(flagWithCronJob.name, flagWithCronJob.shortName, flagWithCronJob.defaultValue.(bool), flagWithCronJob.usage)
	getFlags(cmd, persistent).BoolP(flagWithWs.name, flagWithWs.shortName, flagWithWs.defaultValue.(bool), flagWithWs.usage)
	getFlags(cmd, persistent).BoolP(flagWithoutHttp.name, flagWithoutHttp.shortName, flagWithoutHttp.defaultValue.(bool), flagWithoutHttp.usage)
	getFlags(cmd, persistent).BoolP(flagWithoutMQ.name, flagWithoutMQ.shortName, flagWithoutMQ.defaultValue.(bool), flagWithoutMQ.usage)
}

// 从命令中获取AppName
func getAppName(cmd *cobra.Command) AppName {
	return AppName(cmd.Flag(flagAppName.name).Value.String())
}

// 从命令中获取Env
func getAppEnvironment(cmd *cobra.Command) Env {
	return Env(cmd.Flag(flagAppEnvironment.name).Value.String())
}

// 从命令中获取Version
func getAppVersion(cmd *cobra.Command) Version {
	return Version(cmd.Flag(flagAppVersion.name).Value.String())
}

// 从命令中获取ConfigPath
func getAppConfigPath(cmd *cobra.Command) ConfigFilePath {
	return ConfigFilePath(cmd.Flag(flagAppConfig.name).Value.String())
}

// 从命令中获取WithCronJob
func getWithCronJob(cmd *cobra.Command) bool {
	var (
		withCronJob bool
		err         error
	)
	if withCronJob, err = cmd.Flags().GetBool(flagWithCronJob.name); err != nil {
		panic(err)
	}
	return withCronJob
}

// 从命令中获取WithMQ
func getWithWs(cmd *cobra.Command) bool {
	var (
		withWs bool
		err    error
	)
	if withWs, err = cmd.Flags().GetBool(flagWithWs.name); err != nil {
		panic(err)
	}
	return withWs
}

// 从命令中获取WithoutHttp
func getWithOutHttp(cmd *cobra.Command) bool {
	var (
		withoutHttp bool
		err         error
	)
	if withoutHttp, err = cmd.Flags().GetBool(flagWithoutHttp.name); err != nil {
		panic(err)
	}
	return withoutHttp
}

// 从命令中获取WithMQ
func getWithoutMQConfig(cmd *cobra.Command) bool {
	var (
		withoutMQ bool
		err       error
	)
	if withoutMQ, err = cmd.Flags().GetBool(flagWithoutHttp.name); err != nil {
		panic(err)
	}
	return withoutMQ
}
