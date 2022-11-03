package cmd

import (
	"fmt"
	"os"
	"text/template"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/tikivn/vampire/tlp"
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
		Name: "type",
		Prompt: &survey.Select{
			Message: "App Type:",
			Options: []string{"API", "Worker"},
			Default: "API",
		},
	}
	PrometheusQuestion = &survey.Question{
		Name:   "prometheus",
		Prompt: &survey.Confirm{Message: "Set up Prometheus client?"},
	}
	DatabaseQuestion = &survey.Question{
		Name: "database",
		Prompt: &survey.Select{
			Message: "Database:",
			Options: []string{"Skip", "MongoDB"},
			Default: "Skip",
		},
	}
	WebFramework = &survey.Question{
		Name: "framework",
		Prompt: &survey.Select{
			Message: "Web Framework:",
			Options: []string{"Gin"},
			Default: "Gin",
		},
	}
	WorkerName = &survey.Question{
		Name:     "workername",
		Prompt:   &survey.Input{Message: "Worker Name:"},
		Validate: survey.Required,
	}
)

type InitAnswer struct {
	Module     string
	Dockerfile bool
	Type       string
	Prometheus bool
	Database   string
	Framework  string
	Workername string
}

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Init Project",
	Run: func(cmd *cobra.Command, args []string) {
		var answers InitAnswer

		cobra.CheckErr(startSurvey(&answers))
		cobra.CheckErr(initProject(answers))
	},
}

func startSurvey(answers *InitAnswer) error {

	if err := survey.Ask([]*survey.Question{ModuleNameQuestion, DockerfileQuestion, DatabaseQuestion, TypeQuestion}, answers); err != nil {
		return err
	}
	switch answers.Type {
	case "API":
		if err := survey.Ask([]*survey.Question{WebFramework, PrometheusQuestion}, answers); err != nil {
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

	// env
	touch("development.env")
	touch("production.env")
	// go get mod

	modList := []string{"github.com/rs/zerolog/log", "github.com/joho/godotenv"}
	if answers.Database == "MongoDB" {
		modList = append(modList, "go.mongodb.org/mongo-driver/mongo")
	}
	if answers.Type == "API" {
		modList = append(modList, "github.com/gin-gonic/gin")
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
		mainTemplate := template.Must(template.New("main").Parse(tlp.MainAPI))
		if err := mainTemplate.Execute(apiFile, struct {
			MongoDB bool
			Gin     bool
		}{answers.Database == "MongoDB", answers.Framework == "Gin"}); err != nil {
			return err
		}
	}

	// create worker
	if answers.Type == "Worker" {
		if _, err = os.Stat(fmt.Sprintf("%s/cmd/%s", absolutePath, answers.Workername)); os.IsNotExist(err) {
			if err := os.MkdirAll(fmt.Sprintf("%s/cmd/%s", absolutePath, answers.Workername), 0751); err != nil {
				return err
			}
		}
		workerFile, err := os.Create(fmt.Sprintf("%s/cmd/%s/main.go", absolutePath, answers.Workername))
		if err != nil {
			return err
		}
		defer workerFile.Close()
		workerTemplate := template.Must(template.New("worker").Parse(tlp.MainWorker))
		if err := workerTemplate.Execute(workerFile, struct {
			MongoDB bool
		}{answers.Database == "MongoDB"}); err != nil {
			return err
		}
	}

	return nil
}
