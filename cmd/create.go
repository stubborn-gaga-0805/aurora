package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"
	"github.com/stubborn-gaga-0805/aurora/consts"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var replaceDirs = []string{"api", "cmd", "configs", "internal", "third_party"}

type createCmd struct {
	*baseCmd
	*createFlags

	projectName string
	projectPath string
	branch      string
	sshPath     string
	workingDir  string

	done chan error
}

type createFlags struct {
	flagProjectPath string
	flagIsDemo      bool
}

var (
	flagProjectPath = flag{"path", "p", "", `项目路径`}
	flagIsDemo      = flag{"with.demo", "", false, `是否创建Demo项目`}
)

func newCreateCmd() *createCmd {
	var create = new(createCmd)
	create.baseCmd = newBaseCmd()
	create.done = make(chan error, 1)
	create.cmd = &cobra.Command{
		Use:     "create",
		Aliases: []string{},
		Short:   "创建一个项目",
		Long:    `创建一个项目, 例如: aurora create my-app`,
		Run: func(cmd *cobra.Command, args []string) {
			create.initCreateRuntime(cmd)
			create.run(args)
		},
	}
	addCreateRuntimeFlag(create.cmd, true)

	return create
}

func (create *createCmd) initCreateRuntime(cmd *cobra.Command) {
	create.id, _ = os.Hostname()
	create.env = Env(os.Getenv(consts.OSEnvKey))
	create.createFlags = &createFlags{
		flagProjectPath: getProjectPath(cmd),
		flagIsDemo:      getIsDemo(cmd),
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	create.sshPath = filepath.Join(homeDir, ".ssh")
	if create.workingDir, err = os.Getwd(); err != nil {
		panic(err)
	}
	return
}

func (create *createCmd) run(args []string) {
	var (
		err         error
		projectName string
		workingDir  = create.workingDir
	)

	// 检查项目名
	if len(args) == 0 {
		promptName := &survey.Input{
			Message: "请输入项目名称:",
		}
		if err = survey.AskOne(promptName, &projectName, survey.WithIcons(func(icons *survey.IconSet) {
			icons.Question.Text = "🛠"
			icons.Question.Format = "blue+b"
			icons.Error.Text = "❌"
		}), survey.WithValidator(survey.Required)); err != nil {
			fmt.Println("🚧 Stopped...")
			return
		}
	} else {
		projectName = args[0]
	}
	// 检查是否指定了路径
	if len(create.flagProjectPath) == 0 {
		promptPath := &survey.Input{
			Message: "请输入项目路径:",
			Default: fmt.Sprintf("默认: %s/%s", workingDir, projectName),
		}
		if err = survey.AskOne(promptPath, &workingDir, survey.WithIcons(func(icons *survey.IconSet) {
			icons.Question.Text = "📁"
			icons.Question.Format = "blue+b"
		})); err != nil {
			fmt.Println("🚧 Stopped...")
			return
		}
		workingDir = strings.ReplaceAll(workingDir, "默认: ", "")
		create.projectPath = filepath.Join(workingDir, projectName)
	}
	create.projectName, create.projectPath = parseProjectParams(projectName, create.flagProjectPath)
	// 是否指定为demo
	create.branch = consts.BranchProject
	if !create.flagIsDemo {
		promptDemo := &survey.Confirm{
			Message: "是否创建Demo?",
			Help:    "Demo项目附带框架示例代码，生产环境请勿创建Demo",
			Default: false,
		}
		if err = survey.AskOne(promptDemo, &create.flagIsDemo, survey.WithIcons(func(icons *survey.IconSet) {
			icons.Question.Text = "💡"
			icons.Question.Format = "blue+b"
		})); err != nil {
			fmt.Println("🚧 Stopped...")
			return
		}
		if create.flagIsDemo {
			create.branch = consts.BranchDemo
		}
	}
	go func() {
		create.done <- create.pullRepo()
	}()
	select {
	case <-create.ctx.Done():
		if errors.Is(create.ctx.Err(), context.DeadlineExceeded) {
			fmt.Fprint(os.Stderr, "\033[31mERROR: project creation timed out\033[m\n")
			return
		}
		fmt.Fprintf(os.Stderr, "\033[31mERROR: failed to create project(%s)\033[m\n", create.ctx.Err().Error())
	case err = <-create.done:
		if err != nil {
			fmt.Fprintf(os.Stderr, "\033[31mERROR: Failed to create project(%s)\033[m\n", err.Error())
		}
	}
	return
}

func (create *createCmd) pullRepo() (err error) {
	var (
		override   bool
		targetPath = filepath.Join(create.projectPath, create.projectName)
	)

	// 目标文件夹已存在
	if _, err = os.Stat(targetPath); !os.IsNotExist(err) {
		err = nil
		fmt.Printf("🤔 [目标路径: %s] 已存在！\n", targetPath)
		prompt := &survey.Confirm{
			Message: "是否覆盖 ?",
			Default: false,
			Help:    "选择覆盖将删除现有目录下所有内容",
		}
		if e := survey.AskOne(prompt, &override, survey.WithIcons(func(icons *survey.IconSet) {
			icons.Question.Text = "📥"
			icons.Question.Format = "blue+b"
		})); e != nil {
			return err
		}
		if !override {
			return errors.New(fmt.Sprintf("🚫 创建项目失败，目标文件夹已存在..."))
		}
		// 清空
		_ = os.RemoveAll(targetPath)
	}
	if err := os.MkdirAll(targetPath, fs.ModePerm); err != nil {
		return err
	}
	fmt.Printf("\n\n🚀 正在创建项目: [%s] [From %s To: %s], 拉取分支[%s], 请稍后...\n", color.GreenString(create.projectName), color.BlueString(consts.GoFrameRepoUrl), color.BlueString(create.projectPath), color.BlueString(create.branch))
	if err = create.cloneRepoWithGit(targetPath); err != nil {
		_ = os.RemoveAll(targetPath)
		return errors.New(fmt.Sprintf("🚫 拉取远程仓库失败，无法创建项目... (err: %v)", err))
	}
	fmt.Printf("\n⚙️ 成功拉取项目，初始化GIT仓库与分支...\n")
	if err = create.processLocalRepo(targetPath); err != nil {
		_ = os.RemoveAll(targetPath)
		return errors.New(fmt.Sprintf("🚫 初始化仓库失败，无法创建项目... (err: %v)", err))
	}
	fmt.Printf("\n⚙️ 初始化GIT仓库与分支成功，初始化go.mod...\n")
	if err = create.processGoMod(); err != nil {
		_ = os.RemoveAll(targetPath)
		return errors.New(fmt.Sprintf("🚫 初始化go.mod错误，无法创建项目... (err: %v)", err))
	}
	fmt.Printf("\n +++++++++ ️🎉🎊 项目 [%s] 创建成功...！🍺🍺🍺 +++++++++\n", color.GreenString(create.projectName))
	fmt.Printf(" 📡 当前本地分支: [%s], 你可以运行命令: %s 关联远程仓库...\n", color.GreenString("main"), color.GreenString("git remote add origin <YourGitRepositoryUrl.git>"))
	fmt.Printf(" 📡 运行命令: %s, 将本地分支提交到远程仓库...\n", color.GreenString("git push -u origin main"))

	return nil
}

func (create *createCmd) cloneRepoWithGit(targetPath string) (err error) {
	cmd := exec.Command(
		"git",
		"clone",
		consts.GoFrameRepoUrl,
		targetPath,
		"-b", create.branch,
		"--depth", "1",
		"--single-branch",
		"--no-tags",
		"--verbose",
	)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (create *createCmd) processLocalRepo(targetPath string) (err error) {
	var repo *git.Repository
	if repo, err = git.PlainOpen(targetPath); err != nil {
		return err
	}
	fmt.Printf("✅️ 断开与远程模板仓库的关联...\n")
	if err = repo.DeleteRemote("origin"); err != nil {
		return err
	}

	fmt.Printf("✅ 初始化本地仓库...\n")
	headRef, err := repo.Head()
	if err != nil {
		return err
	}
	branchName := headRef.Name().Short()
	fmt.Printf("✅️ 当前分支: [%s]\n", color.BlueString(branchName))

	// 将分支自改为main
	if branchName != consts.BranchMain {
		branchRef, err := repo.Reference(plumbing.ReferenceName("refs/heads/"+branchName), true)
		if err != nil {
			return err
		}
		oldBranchRef := plumbing.NewHashReference(plumbing.ReferenceName("refs/heads/"+branchName), branchRef.Hash())
		newBranchRef := plumbing.NewHashReference("refs/heads/"+consts.BranchMain, oldBranchRef.Hash())

		fmt.Printf("✅ 当前分支不为[%s], 创建: [%s]分支\n", color.GreenString(consts.BranchMain), color.GreenString(consts.BranchMain))
		if err = repo.Storer.SetReference(newBranchRef); err != nil {
			return err
		}
		fmt.Printf("✅️ 删除: [%s]分支\n", color.GreenString(branchName))
		if err = repo.Storer.RemoveReference(oldBranchRef.Name()); err != nil {
			return err
		}
		// 切换到main分支
		mainRef, err := repo.Reference("refs/heads/"+consts.BranchMain, true)
		if err != nil {
			return err
		}
		workTree, err := repo.Worktree()
		if err != nil {
			return err
		}
		if err = workTree.Checkout(&git.CheckoutOptions{
			Branch: mainRef.Name(),
		}); err != nil {
			return err
		}
	}

	return nil
}

func (create *createCmd) processGoMod() (err error) {
	var cmd = new(exec.Cmd)
	// 设置包名为项目名
	cmd = exec.Command(
		"go",
		"mod",
		"edit",
		"-module",
		fmt.Sprintf("%s", create.projectName),
	)
	cmd.Dir = fmt.Sprintf("%s/%s", create.projectPath, create.projectName)
	if err := cmd.Run(); err != nil {
		return err
	}
	fmt.Printf("✅️ 设置go.mod的模块名为[%s]\n", color.BlueString(create.projectName))

	// 查找并修改引用
	if err = replaceImport(create.projectName, filepath.Join(create.projectPath, create.projectName)); err != nil {
		return err
	}

	// tidy
	cmd = exec.Command(
		"go",
		"mod",
		"tidy",
	)
	cmd.Dir = fmt.Sprintf("%s/%s", create.projectPath, create.projectName)
	if err := cmd.Run(); err != nil {
		return err
	}
	fmt.Printf("✅️ go mod tidy...\n")
	return nil
}

func addCreateRuntimeFlag(cmd *cobra.Command, persistent bool) {
	getFlags(cmd, persistent).StringP(flagProjectPath.name, flagProjectPath.shortName, flagProjectPath.defaultValue.(string), flagProjectPath.usage)
	getFlags(cmd, persistent).BoolP(flagIsDemo.name, flagIsDemo.shortName, flagIsDemo.defaultValue.(bool), flagIsDemo.usage)
}

func getProjectPath(cmd *cobra.Command) string {
	return cmd.Flag(flagProjectPath.name).Value.String()
}

func getIsDemo(cmd *cobra.Command) bool {
	var (
		isDemo bool
		err    error
	)
	if isDemo, err = cmd.Flags().GetBool(flagIsDemo.name); err != nil {
		panic(err)
	}
	return isDemo
}

func parseProjectParams(projectName string, workingDir string) (projectNameResult, workingDirResult string) {
	_projectDir := projectName
	_workingDir := workingDir
	if strings.HasPrefix(projectName, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// cannot get user home return fallback place dir
			return _projectDir, _workingDir
		}
		_projectDir = filepath.Join(homeDir, projectName[2:])
	}
	if !filepath.IsAbs(workingDir) {
		absPath, err := filepath.Abs(workingDir)
		if err != nil {
			return _projectDir, _workingDir
		}
		_projectDir = filepath.Join(absPath, _projectDir)
	}

	return filepath.Base(_projectDir), filepath.Dir(_projectDir)
}

func replaceImport(moduleName, workdir string) (err error) {
	// 替换main文件
	mainPath := filepath.Join(workdir, "main.go")
	content, err := ioutil.ReadFile(mainPath)
	if err != nil {
		return err
	}
	newContent := strings.ReplaceAll(string(content), consts.GoFrameModule, moduleName)
	err = ioutil.WriteFile(mainPath, []byte(newContent), 0644)
	if err != nil {
		return err
	}
	fmt.Printf("💡 替换文件: %s 中的引用为: %s\n", mainPath, moduleName)

	// 替换文件夹中的文件
	for _, dir := range replaceDirs {
		err = filepath.Walk(filepath.Join(workdir, dir), func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				// 读取文件内容
				content, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}

				// 查找并替换字符串
				newContent := strings.ReplaceAll(string(content), consts.GoFrameModule, moduleName)

				// 将替换后的内容写回文件
				err = ioutil.WriteFile(path, []byte(newContent), info.Mode())
				if err != nil {
					return err
				}
				fmt.Printf("💡 替换文件: %s 中的引用为: %s\n", path, color.GreenString(fmt.Sprintf("module: %s", moduleName)))
			}
			return nil
		})
	}

	return err
}
