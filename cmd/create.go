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
	flagProjectPath = flag{"path", "p", "", `project path`}
	flagIsDemo      = flag{"with.demo", "", false, `whether to create a 'demo' project`}
)

func newCreateCmd() *createCmd {
	var create = new(createCmd)
	create.baseCmd = newBaseCmd()
	create.done = make(chan error, 1)
	create.cmd = &cobra.Command{
		Use:     "create",
		Aliases: []string{},
		Short:   "create a new project",
		Long:    `create a new project, eg: aurora create my-app`,
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
			Message: "enter a name for the project:",
		}
		if err = survey.AskOne(promptName, &projectName, survey.WithIcons(func(icons *survey.IconSet) {
			icons.Question.Text = "🛠"
			icons.Question.Format = "blue+b"
			icons.Error.Text = "❌"
		}), survey.WithValidator(survey.Required)); err != nil {
			fmt.Println("🚧 Stopped...something went wrong")
			return
		}
	} else {
		projectName = args[0]
	}
	// 检查是否指定了路径
	if len(create.flagProjectPath) == 0 {
		promptPath := &survey.Input{
			Message: "Please enter the project path:",
			Default: fmt.Sprintf("default: %s/%s", workingDir, projectName),
		}
		if err = survey.AskOne(promptPath, &workingDir, survey.WithIcons(func(icons *survey.IconSet) {
			icons.Question.Text = "📁"
			icons.Question.Format = "blue+b"
		})); err != nil {
			fmt.Println("🚧 Stopped...something went wrong")
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
			Message: "Whether to create 'demo' code?",
			Help:    "The Demo project comes with framework sample code, please do not create a Demo in the production environment",
			Default: false,
		}
		if err = survey.AskOne(promptDemo, &create.flagIsDemo, survey.WithIcons(func(icons *survey.IconSet) {
			icons.Question.Text = "💡"
			icons.Question.Format = "blue+b"
		})); err != nil {
			fmt.Println("🚧 Stopped...something went wrong")
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
		fmt.Printf("🤔 [Target path: %s] already exists！\n", targetPath)
		prompt := &survey.Confirm{
			Message: "Whether to overwrite existing directories ?",
			Default: false,
			Help:    "WARNING: Selecting overwrite will delete all content under the existing directory",
		}
		if e := survey.AskOne(prompt, &override, survey.WithIcons(func(icons *survey.IconSet) {
			icons.Question.Text = "📥"
			icons.Question.Format = "blue+b"
		})); e != nil {
			return err
		}
		if !override {
			return errors.New(fmt.Sprintf("🚫 Failed to create project, target folder already exists..."))
		}
		// 清空
		_ = os.RemoveAll(targetPath)
	}
	if err := os.MkdirAll(targetPath, fs.ModePerm); err != nil {
		return err
	}
	fmt.Printf("\n\n🚀 Creating project: [%s] [From %s To: %s], Pulling GIT branch[%s], please wait...\n", color.GreenString(create.projectName), color.BlueString(consts.GoFrameRepoUrl), color.BlueString(create.projectPath), color.BlueString(create.branch))
	if err = create.cloneRepoWithGit(targetPath); err != nil {
		_ = os.RemoveAll(targetPath)
		return errors.New(fmt.Sprintf("🚫 Failed to pull the remote git repostory, unable to create the project... (err: %v)", err))
	}
	fmt.Printf("\n⚙️ Successfully pulled project, initializing GIT repository and branch...\n")
	if err = create.processLocalRepo(targetPath); err != nil {
		_ = os.RemoveAll(targetPath)
		return errors.New(fmt.Sprintf("🚫 Failed to initialize GIT repository, unable to create project... (err: %v)", err))
	}
	fmt.Printf("\n⚙️ Initializing GIT repository and branch succeeded ! initializing go.mod file...\n")
	if err = create.processGoMod(); err != nil {
		_ = os.RemoveAll(targetPath)
		return errors.New(fmt.Sprintf("🚫 There was an error initializing the go.mod file and the project could not be created... (err: %v)", err))
	}
	fmt.Printf("\n +++++++++ ️🎉🎊 Project [%s] created successfully...！🍺🍺🍺 +++++++++\n", color.GreenString(create.projectName))
	fmt.Printf(" 📡 Current local GIT branch: [%s], You can run the command: %s to associate a remote repository...\n", color.GreenString("main"), color.GreenString("git remote add origin <YourGitRepositoryUrl.git>"))
	fmt.Printf(" 📡 You can run the command: %s, to push your local GIT branch to the remote repository...\n", color.GreenString("git push -u origin main"))
	fmt.Printf(" 🍻 All processes are successful! Enjoy the fun of coding...🥳\n")

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
	fmt.Printf("✅️ Disassociate from remote GIT template repository...\n")
	if err = repo.DeleteRemote("origin"); err != nil {
		return err
	}

	fmt.Printf("✅ Initialize the local GIT repository...\n")
	headRef, err := repo.Head()
	if err != nil {
		return err
	}
	branchName := headRef.Name().Short()
	fmt.Printf("✅️ current GIT branch: [%s]\n", color.BlueString(branchName))

	// 将分支自改为main
	if branchName != consts.BranchMain {
		branchRef, err := repo.Reference(plumbing.ReferenceName("refs/heads/"+branchName), true)
		if err != nil {
			return err
		}
		oldBranchRef := plumbing.NewHashReference(plumbing.ReferenceName("refs/heads/"+branchName), branchRef.Hash())
		newBranchRef := plumbing.NewHashReference("refs/heads/"+consts.BranchMain, oldBranchRef.Hash())

		fmt.Printf("✅ The current GIT branch is not[%s], create: [%s]branch\n", color.GreenString(consts.BranchMain), color.GreenString(consts.BranchMain))
		if err = repo.Storer.SetReference(newBranchRef); err != nil {
			return err
		}
		fmt.Printf("✅️ delete branch: [%s]\n", color.GreenString(branchName))
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
	fmt.Printf("✅️ Set the module name of 'go.mod' to [%s]\n", color.BlueString(create.projectName))

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
	fmt.Printf("💡 Replace references in file: %s with: %s\n", mainPath, moduleName)

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
				fmt.Printf("💡 Replace references in file: %s with: %s\n", path, color.GreenString(fmt.Sprintf("module: %s", moduleName)))
			}
			return nil
		})
	}

	return err
}
