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
	flagAppName        = flag{"name", "n", "prepare-to-go", "设置服务名称"}
	flagAppEnvironment = flag{"env", "e", "local", "设置服务的运行环境"}
	flagAppVersion     = flag{"version", "v", "v1.0", "设置应用的版本"}
	flagAppConfig      = flag{"config", "c", "", "设置配置文件的路径"}
	flagWithCronJob    = flag{"with.cron", "", false, "是否启动定时任务"}
	flagWithWs         = flag{"with.ws", "", false, "是否启动websocket服务"}
	flagWithoutHttp    = flag{"without.server", "", false, "不启动http服务"}
	flagWithoutMQ      = flag{"without.mq", "", false, "不启动MQ"}
)

func newRunCmd() *runCmd {
	run := &runCmd{newBaseCmd(), new(runFlags)}
	run.cmd = &cobra.Command{
		Use:     "run",
		Aliases: []string{"start", "running", "up"},
		Short:   "启动web服务 (如：http, grpc, websocket)，默认启动Http服务",
		Long:    `例如: aurora run --with.cron`,
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
		fmt.Println("🚫 当前目录下没有找到main文件，请在项目根目录运行...")
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
		panic(fmt.Sprintf("不支持的运行环境... 【%s】", run.env))
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
	fmt.Printf("+++++++++ %+v", run.runFlags)
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
