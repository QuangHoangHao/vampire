package cmd

import (
	"fmt"
	"os"
	"text/template"

	"github.com/AlecAivazis/survey/v2"
	"github.com/QuangHoangHao/vampire/tlp"
	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
)

var (
	ModuleNameQuestion = &survey.Question{
		Name:     "module",
		Prompt:   &survey.Input{Message: "Module name:"},
		Validate: survey.Required,
	}
	DockerfileQuestion = &survey.Question{
		Name:   "dockerfile",
		Prompt: &survey.Confirm{Message: "Generate Dockerfile?"},
	}
	TypeQuestion = &survey.Question{
		Name:     "type",
		Prompt:   &survey.Select{Message: "Choose type:", Options: []string{"API", "Worker"}},
		Validate: survey.Required,
	}
	WorkerName = &survey.Question{
		Name:     "workerName",
		Prompt:   &survey.Input{Message: "Worker Name:"},
		Validate: survey.Required,
	}
	ControllerName = &survey.Question{
		Name:     "controllerName",
		Prompt:   &survey.Input{Message: "Controller Name:"},
		Validate: survey.Required,
	}
	ServiceName = &survey.Question{
		Name:     "serviceName",
		Prompt:   &survey.Input{Message: "Service Name:"},
		Validate: survey.Required,
	}
	RepoName = &survey.Question{
		Name:     "RepositoryName",
		Prompt:   &survey.Input{Message: "Repository Name:"},
		Validate: survey.Required,
	}
	DatabaseQuestion = &survey.Question{
		Name: "database",
		Prompt: &survey.Select{
			Message: "Database:",
			Options: []string{"Skip", "MongoDB"},
			Default: "Skip",
		},
	}
	DINameQuestion = &survey.Question{
		Name:     "DIName",
		Prompt:   &survey.Input{Message: "DIName Name:"},
		Validate: survey.Required,
	}
)

type InitAnswer struct {
	Module     string
	Type       string
	Dockerfile bool
	WorkerName string
}

type DIAnswer struct {
	Type     string
	DIName   string
	Database string
}

type WorkerAnswer struct {
	WorkerName string
}

type ControllerAnswer struct {
	ControllerName string
}

type ServiceAnswer struct {
	ServiceName string
}

type RepositoryAnswer struct {
	RepositoryName string
	Database       string
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(workerCmd)
	rootCmd.AddCommand(controllerCmd)
	rootCmd.AddCommand(serviceCmd)
	rootCmd.AddCommand(repoCmd)
	rootCmd.AddCommand(DICmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Init Project",
	Run: func(cmd *cobra.Command, args []string) {
		var answers InitAnswer

		cobra.CheckErr(startSurveyInit(&answers))
		cobra.CheckErr(initProject(answers))
	},
}

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Create Worker",
	Run: func(cmd *cobra.Command, args []string) {
		var answers WorkerAnswer

		cobra.CheckErr(startSurveyWorker(&answers))
		cobra.CheckErr(createWorker(answers.WorkerName))
	},
}

var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "Create Controller",
	Run: func(cmd *cobra.Command, args []string) {
		var answers ControllerAnswer

		cobra.CheckErr(startSurveyController(&answers))
		cobra.CheckErr(createController(answers.ControllerName))
	},
}

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Create Service",
	Run: func(cmd *cobra.Command, args []string) {
		var answers ServiceAnswer

		cobra.CheckErr(startSurveyService(&answers))
		cobra.CheckErr(createService(answers.ServiceName))
	},
}

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Create Repo",
	Run: func(cmd *cobra.Command, args []string) {
		var answers RepositoryAnswer

		cobra.CheckErr(startSurveyRepo(&answers))
		cobra.CheckErr(createRepo(answers))
	},
}

var DICmd = &cobra.Command{
	Use:   "DI",
	Short: "Create DI",
	Run: func(cmd *cobra.Command, args []string) {
		var answers DIAnswer

		cobra.CheckErr(startSurveyDI(&answers))
		cobra.CheckErr(createDI(answers))
	},
}

func startSurveyInit(answers *InitAnswer) error {
	if err := survey.Ask([]*survey.Question{ModuleNameQuestion, DockerfileQuestion, TypeQuestion}, answers); err != nil {
		return err
	}
	switch answers.Type {
	case "API":
		if err := survey.Ask([]*survey.Question{}, answers); err != nil {
			return err
		}
	case "Worker":
		if err := survey.Ask([]*survey.Question{WorkerName}, answers); err != nil {
			return err
		}
	}
	return nil
}

func initProject(answers InitAnswer) error {
	absolutePath, err := os.Getwd()

	if err != nil {
		return err
	}

	// init go.mod
	if err := goInit(answers.Module); err != nil {
		return err
	}

	// create README
	readmeFile, err := os.Create(fmt.Sprintf("%s/README.md", absolutePath))
	if err != nil {
		return err
	}
	defer readmeFile.Close()
	readmeTemplate := template.Must(template.New("README").Parse(tlp.ReadmeTemplate))
	if err := readmeTemplate.Execute(readmeFile, nil); err != nil {
		return err
	}

	// init git, .gitignore
	if err := gitInit(); err != nil {
		return err
	}
	gitignore, err := os.Create(fmt.Sprintf("%s/.gitignore", absolutePath))
	if err != nil {
		return err
	}
	defer gitignore.Close()
	gitignoreTemplate := template.Must(template.New("gitignore").Parse(tlp.Gitignore))
	if err := gitignoreTemplate.Execute(gitignore, nil); err != nil {
		return err
	}

	// dockerfile
	if answers.Dockerfile {
		dockerfile, err := os.Create(fmt.Sprintf("%s/Dockerfile", absolutePath))
		if err != nil {
			return err
		}
		defer dockerfile.Close()
		DockerTemplate := template.Must(template.New("dockerfile").Parse(tlp.Dockerfile))
		if err := DockerTemplate.Execute(dockerfile, nil); err != nil {
			return err
		}
	}

	// create env
	touch("development.env")
	touch("production.env")

	envFile, err := os.OpenFile(fmt.Sprintf("%s/development.env", absolutePath), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer envFile.Close()

	envTemplate := template.Must(template.New("env").Parse(tlp.Env))
	if err := envTemplate.Execute(envFile, nil); err != nil {
		return err
	}

	// go get mod
	modList := []string{"github.com/rs/zerolog/log", "github.com/joho/godotenv"}
	if answers.Type == "API" {
		modList = append(modList, "github.com/gin-gonic/gin")
	}
	if answers.Type == "Worker" {
		modList = append(modList, "github.com/segmentio/kafka-go", "github.com/Shopify/sarama", "github.com/QuangHoangHao/kafka-go")
	}

	for _, mod := range modList {
		if err := goGet(mod); err != nil {
			return err
		}
	}

	// create cmd/api/main.go
	if answers.Type == "API" {
		if _, err = os.Stat(fmt.Sprintf("%s/cmd/api", absolutePath)); os.IsNotExist(err) {
			if err := os.MkdirAll(fmt.Sprintf("%s/cmd/api", absolutePath), 0751); err != nil {
				return err
			}
		}
		apiFile, err := os.Create(fmt.Sprintf("%s/cmd/api/main.go", absolutePath))
		if err != nil {
			return err
		}
		defer apiFile.Close()
		apiTemplate := template.Must(template.New("api").Parse(tlp.API))
		if err := apiTemplate.Execute(apiFile, nil); err != nil {
			return err
		}

		return nil
	}

	// create cmd/worker/main.go
	if answers.Type == "Worker" {
		if err = createWorker(answers.WorkerName); err != nil {
			return err
		}
	}

	return nil
}

func startSurveyWorker(answers *WorkerAnswer) error {
	if err := survey.Ask([]*survey.Question{WorkerName}, answers); err != nil {
		return err
	}
	return nil
}

func createWorker(workerName string) error {
	absolutePath, err := os.Getwd()
	if err != nil {
		return err
	}

	funcMap := template.FuncMap{
		"ToCamel":      strcase.ToCamel,
		"ToLowerCamel": strcase.ToLowerCamel,
	}

	if _, err = os.Stat(fmt.Sprintf("%s/cmd/worker/%s", absolutePath, workerName)); os.IsNotExist(err) {
		if err := os.MkdirAll(fmt.Sprintf("%s/cmd/worker/%s", absolutePath, workerName), 0751); err != nil {
			return err
		}
	}
	workerFile, err := os.Create(fmt.Sprintf("%s/cmd/worker/%s/main.go", absolutePath, workerName))
	if err != nil {
		return err
	}
	defer workerFile.Close()

	workerTemplate := template.Must(template.New("worker").Parse(tlp.Worker))
	if err := workerTemplate.Execute(workerFile, struct {
		WorkerName string
	}{workerName}); err != nil {
		return err
	}

	if _, err = os.Stat(fmt.Sprintf("%s/internal/%s", absolutePath, workerName)); os.IsNotExist(err) {
		if err := os.MkdirAll(fmt.Sprintf("%s/internal/%s", absolutePath, workerName), 0751); err != nil {
			return err
		}
	}
	workerHandlerFile, err := os.Create(fmt.Sprintf("%s/internal/%s/%s.handler.go", absolutePath, workerName, strcase.ToKebab(workerName)))
	if err != nil {
		return err
	}

	defer workerHandlerFile.Close()

	workerHandlerTemplate := template.Must(template.New("workerHandler").Funcs(funcMap).Parse(tlp.WorkerHandler))
	if err := workerHandlerTemplate.Execute(workerHandlerFile, struct {
		Module     string
		WorkerName string
	}{absolutePath, workerName}); err != nil {
		return err
	}

	return nil
}

func startSurveyController(answers *ControllerAnswer) error {
	if err := survey.Ask([]*survey.Question{ControllerName}, answers); err != nil {
		return err
	}
	return nil
}

func createController(controllerName string) error {
	absolutePath, err := os.Getwd()
	if err != nil {
		return err
	}

	funcMap := template.FuncMap{
		"ToCamel":      strcase.ToCamel,
		"ToLowerCamel": strcase.ToLowerCamel,
	}

	if _, err = os.Stat(fmt.Sprintf("%s/internal/%s", absolutePath, controllerName)); os.IsNotExist(err) {
		if err := os.MkdirAll(fmt.Sprintf("%s/internal/%s", absolutePath, controllerName), 0751); err != nil {
			return err
		}
	}
	controllerFile, err := os.Create(fmt.Sprintf("%s/internal/%s/%s.controller.go", absolutePath, controllerName, strcase.ToKebab(controllerName)))
	if err != nil {
		return err
	}
	defer controllerFile.Close()

	controllerTemplate := template.Must(template.New("controller").Funcs(funcMap).Parse(tlp.Controller))
	if err := controllerTemplate.Execute(controllerFile, struct {
		Name string
	}{controllerName}); err != nil {
		return err
	}

	return nil
}

func startSurveyService(answers *ServiceAnswer) error {
	if err := survey.Ask([]*survey.Question{ServiceName}, answers); err != nil {
		return err
	}
	return nil
}

func createService(serviceName string) error {
	absolutePath, err := os.Getwd()
	if err != nil {
		return err
	}

	funcMap := template.FuncMap{
		"ToCamel":      strcase.ToCamel,
		"ToLowerCamel": strcase.ToLowerCamel,
	}

	if _, err = os.Stat(fmt.Sprintf("%s/internal/%s", absolutePath, serviceName)); os.IsNotExist(err) {
		if err := os.MkdirAll(fmt.Sprintf("%s/internal/%s", absolutePath, serviceName), 0751); err != nil {
			return err
		}
	}

	serviceFile, err := os.Create(fmt.Sprintf("%s/internal/%s/%s.service.go", absolutePath, serviceName, strcase.ToKebab(serviceName)))
	if err != nil {
		return err
	}

	defer serviceFile.Close()

	serviceTemplate := template.Must(template.New("service").Funcs(funcMap).Parse(tlp.Service))

	if err := serviceTemplate.Execute(serviceFile, struct {
		Name string
	}{serviceName}); err != nil {
		return err
	}

	return nil
}

func startSurveyRepo(answers *RepositoryAnswer) error {
	if err := survey.Ask([]*survey.Question{DatabaseQuestion, RepoName}, answers); err != nil {
		return err
	}
	return nil
}

func createRepo(answers RepositoryAnswer) error {
	absolutePath, err := os.Getwd()
	if err != nil {
		return err
	}

	funcMap := template.FuncMap{
		"ToCamel":      strcase.ToCamel,
		"ToLowerCamel": strcase.ToLowerCamel,
		"ToKebab":      strcase.ToKebab,
	}

	if _, err = os.Stat(fmt.Sprintf("%s/internal/%s", absolutePath, answers.RepositoryName)); os.IsNotExist(err) {
		if err := os.MkdirAll(fmt.Sprintf("%s/internal/%s", absolutePath, answers.RepositoryName), 0751); err != nil {
			return err
		}
	}

	repoFile, err := os.Create(fmt.Sprintf("%s/internal/%s/%s.repo.go", absolutePath, answers.RepositoryName, strcase.ToKebab(answers.RepositoryName)))
	if err != nil {
		return err
	}

	defer repoFile.Close()

	repoTemplate := template.Must(template.New("repo").Funcs(funcMap).Parse(tlp.Repository))

	if err := repoTemplate.Execute(repoFile, struct {
		MongoDB bool
		Name    string
	}{answers.Database == "MongoDB", answers.RepositoryName}); err != nil {
		return err
	}

	entityFile, err := os.Create(fmt.Sprintf("%s/internal/%s/%s.entity.go", absolutePath, answers.RepositoryName, strcase.ToKebab(answers.RepositoryName)))
	if err != nil {
		return err
	}

	defer entityFile.Close()

	entityTemplate := template.Must(template.New("entity").Funcs(funcMap).Parse(tlp.Entity))

	if err := entityTemplate.Execute(entityFile, struct {
		MongoDB bool
		Name    string
	}{answers.Database == "MongoDB", answers.RepositoryName}); err != nil {
		return err
	}

	return nil
}

func startSurveyDI(answers *DIAnswer) error {
	if err := survey.Ask([]*survey.Question{TypeQuestion, DINameQuestion, DatabaseQuestion}, answers); err != nil {
		return err
	}
	return nil
}

func createDI(answers DIAnswer) error {
	if answers.Type == "API" {
		if err := createController(answers.DIName); err != nil {
			return err
		}
	}

	if err := createService(answers.DIName); err != nil {
		return err
	}

	if err := createRepo(RepositoryAnswer{answers.DIName, answers.Database}); err != nil {
		return err
	}

	return nil
}
