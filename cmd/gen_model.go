package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stubborn-gaga-0805/aurora/conf"
	"github.com/stubborn-gaga-0805/aurora/consts"
	"github.com/stubborn-gaga-0805/aurora/pkg/mysql"
	"os"
	"os/exec"
	"strings"
)

const (
	defaultOutputPath  = "./internal/repo/orm"
	defaultPackageName = "orm"
	defaultDbConn      = "db"
)

type genModelCmd struct {
	*baseCmd
	*genModelFlags

	conn         []conf.DB
	chooseConn   conf.DB
	chooseTables []string
}

type genModelFlags struct {
	flagTables      string
	flagOutputPath  string
	flagPackageName string
	flagDBConn      string
}

var (
	flagTables      = flag{"table", "t", "", `指定生成的表名(多张表用","隔开)... `}
	flagOutputPath  = flag{"output", "o", defaultOutputPath, `执行生成文件的路径,默认"./internal/repo/orm"... `}
	flagPackageName = flag{"pkg", "p", defaultPackageName, `生成model文件的包名,默认"orm",需要和生成路径的文件夹对应... `}
	flagDBConn      = flag{"conn", "c", defaultDbConn, `配置文件中的连接配置，默认"db"... `}

	dnsTpl = `%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local`
)

func newGenModelCmd() *genModelCmd {
	gen := new(genModelCmd)
	gen.baseCmd = newBaseCmd()
	gen.genModelFlags = new(genModelFlags)
	gen.conn = make([]conf.DB, 0)
	gen.cmd = &cobra.Command{
		Use:     "gen-model",
		Aliases: []string{"model"},
		Short:   "Generate 'model' files for 'gorm'",
		Long:    `💡 Generate 'model' files for 'gorm', eg: aurora gen-model, enter interactive mode`,
		Run: func(cmd *cobra.Command, args []string) {
			gen.initJobRuntime(cmd)
			gen.initConfig()
			gen.run()
		},
	}
	addGenModelRuntimeFlag(gen.cmd, true)

	return gen
}

func (gen *genModelCmd) initJobRuntime(cmd *cobra.Command) {
	// 检查是否在项目目录下
	if !gen.InProjectPath() {
		fmt.Println("🚫 The 'main.go' file is not found in the current directory, please run it in the project root directory...")
		os.Exit(1)
		return
	}

	gen.id, _ = os.Hostname()
	gen.env = Env(os.Getenv(consts.OSEnvKey))
	gen.configFilePath = fmt.Sprintf("./configs/config.%s.yaml", gen.env)
	gen.genModelFlags = &genModelFlags{
		flagTables:      getTables(cmd),
		flagPackageName: getPackageName(cmd),
		flagOutputPath:  getOutputPath(cmd),
		flagDBConn:      getDB(cmd),
	}
	return
}

func (gen *genModelCmd) run() {
	var (
		err error
	)
	if err = gen.parseConfigFile(); err != nil {
		fmt.Printf("🚫[Command: %s] execution failed...[%v]\n", gen.cmd.Use, err)
		return
	}

	// 根据配置文件检查当前项目下存在几个数据库连接
	if len(gen.conn) == 0 {
		fmt.Printf("🚧 It is not detected that your current project has a 'DB' configuration, and the 'model' file cannot be generated...\n")
		return
	}
	if len(gen.conn) == 1 {
		gen.chooseConn = gen.conn[0]
	} else {
		if gen.chooseConn, err = gen.chooseUrDB(); err != nil {
			fmt.Printf("🚫[Command: %s] execution failed...[%v]\n", gen.cmd.Use, err)
			return
		}
	}
	// 判断有没有设置表
	if len(gen.flagTables) != 0 {
		gen.chooseTables = strings.Split(gen.flagTables, ",")
	} else {
		if err = gen.chooseUrTables(); err != nil {
			fmt.Printf("🚫[Command: %s] execution failed...[%v]\n", gen.cmd.Use, err)
			return
		}
	}
	// 生成model文件
	if err := gen.genModelProcess(); err != nil {
		fmt.Printf("🚫[Command: %s] execution failed...[%v]\n", gen.cmd.Use, err)
		return
	}

	fmt.Printf("\n\n🪄🎉🎊 The 'model' file has been generated successfully...😄!\n")

	return
}

func (gen *genModelCmd) parseConfigFile() (err error) {
	viper.SetConfigType("yaml")
	viper.SetConfigFile(gen.configFilePath)
	for k, _ := range viper.Sub("data").AllSettings() {
		if viper.Sub("data").Sub(k).IsSet("driver") {
			var conn conf.DB
			if err = viper.Sub("data").Sub(k).Unmarshal(&conn); err != nil {
				return err
			}
			gen.conn = append(gen.conn, conn)
		}
	}
	return
}

func (gen *genModelCmd) chooseUrDB() (db conf.DB, err error) {
	var (
		chooseDB    string
		selectList  = make([]string, len(gen.conn))
		connMapping = make(map[string]conf.DB, len(gen.conn))
	)
	for i, v := range gen.conn {
		selectList[i] = fmt.Sprintf("[%s: %s]", v.Driver, v.Database)
		connMapping[selectList[i]] = v
	}
	prompt := &survey.Select{
		Message: "It is detected that you have multiple 'DB' connection configurations, please select the 'DB' connection to operate...🤔:",
		Options: selectList,
		Default: selectList[0],
	}
	if err := survey.AskOne(prompt, &chooseDB, survey.WithIcons(func(icons *survey.IconSet) {
		icons.Question.Text = "💿"
		icons.Question.Format = "green+b"
		icons.Help.Format = "green+b"
	}), survey.WithValidator(survey.Required)); err != nil {
		return db, errors.New("🚧 Stopped...something went wrong")
	}
	db, ok := connMapping[chooseDB]
	if !ok {
		return db, errors.New("choose DB Error")
	}
	fmt.Printf("✅ You selected [%s], connecting...\n", chooseDB)
	return db, nil
}

func (gen *genModelCmd) chooseUrTables() (err error) {
	var allTables []string
	db, err := mysql.New(gen.ctx, gen.chooseConn)
	if err != nil {
		return err
	}
	result := db.Raw("SHOW TABLES").Scan(&allTables)
	if result.Error != nil {
		return err
	}
	if len(allTables) == 0 {
		return errors.New("❌ No tables found in the current database")
	}
	prompt := &survey.MultiSelect{
		Message:  "select the table to be generated...😏",
		Options:  allTables,
		PageSize: 15,
	}
	if err = survey.AskOne(prompt, &gen.chooseTables, survey.WithIcons(func(icons *survey.IconSet) {
		icons.Question.Text = "📊"
		icons.Question.Format = "green+b"
		icons.Help.Format = "green+b"
	}), survey.WithKeepFilter(true), survey.WithValidator(survey.Required)); err != nil {
		return errors.New("🚧 Stopped")
	}
	fmt.Printf("✅ You selected %s tables: [%s], The 'model' file is being generated...\n", color.BlueString("%d", len(gen.chooseTables)), color.BlueString(strings.Join(gen.chooseTables, ", ")))

	return nil
}

func (gen *genModelCmd) genModelProcess() error {
	var dns = fmt.Sprintf(dnsTpl, gen.chooseConn.Username, gen.chooseConn.Password, gen.chooseConn.Addr, gen.chooseConn.Database)
	// 组装gen-tools命令
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	command := exec.Command("gentool",
		"-dsn", dns,
		"-db", "mysql",
		"-tables", strings.Join(gen.chooseTables, ","),
		"-modelPkgName", gen.flagPackageName,
		"-outPath", gen.flagOutputPath,
		"-outFile", "model.go",
		"-onlyModel",
		"-fieldWithIndexTag",
		"-fieldWithTypeTag",
		"-fieldNullable",
	)
	command.Dir = wd
	command.Env = os.Environ()

	stdout, _ := command.StdoutPipe()
	stderr, _ := command.StderrPipe()
	if err := command.Start(); err != nil {
		return err
	}
	// 读取命令的标准输出和错误输出
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		fmt.Println(color.RedString(scanner.Text()))
	}
	scanner = bufio.NewScanner(stderr)
	for scanner.Scan() {
		fmt.Println(color.GreenString(scanner.Text()))
	}

	if err := command.Wait(); err != nil {
		return err
	}
	return nil
}

func addGenModelRuntimeFlag(cmd *cobra.Command, persistent bool) {
	getFlags(cmd, persistent).StringP(flagTables.name, flagTables.shortName, flagTables.defaultValue.(string), flagTables.usage)
	getFlags(cmd, persistent).StringP(flagOutputPath.name, flagOutputPath.shortName, flagOutputPath.defaultValue.(string), flagOutputPath.usage)
	getFlags(cmd, persistent).StringP(flagPackageName.name, flagPackageName.shortName, flagPackageName.defaultValue.(string), flagPackageName.usage)
	getFlags(cmd, persistent).StringP(flagDBConn.name, flagDBConn.shortName, flagDBConn.defaultValue.(string), flagDBConn.usage)
}

func getTables(cmd *cobra.Command) string {
	return cmd.Flag(flagTables.name).Value.String()
}

func getOutputPath(cmd *cobra.Command) string {
	return cmd.Flag(flagOutputPath.name).Value.String()
}

func getPackageName(cmd *cobra.Command) string {
	return cmd.Flag(flagPackageName.name).Value.String()
}

func getDB(cmd *cobra.Command) string {
	return cmd.Flag(flagDBConn.name).Value.String()
}
