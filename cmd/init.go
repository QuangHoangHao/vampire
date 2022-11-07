package cmd

import (
	"fmt"
	"os"
	"text/template"

	"github.com/AlecAivazis/survey/v2"
	"github.com/iancoleman/strcase"
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
	Mq = &survey.Question{
		Name: "mq",
		Prompt: &survey.Select{
			Message: "Message Queue:",
			Options: []string{"Skip", "Kafka"},
			Default: "Skip",
		},
	}
	KafkaURL = &survey.Question{
		Name:     "kafkaURL",
		Prompt:   &survey.Input{Message: "Kafka URL:"},
		Validate: survey.Required,
	}
	DBUrl = &survey.Question{
		Name:     "dbURL",
		Prompt:   &survey.Input{Message: "Database URL:"},
		Validate: survey.Required,
	}
	DBName = &survey.Question{
		Name:     "dbName",
		Prompt:   &survey.Input{Message: "Database Name:"},
		Validate: survey.Required,
	}
	NameWorker = &survey.Question{
		Name:     "nameWorker",
		Prompt:   &survey.Input{Message: "Worker Name:"},
		Validate: survey.Required,
	}
	TopicNameWorker = &survey.Question{
		Name:     "topicNameWorker",
		Prompt:   &survey.Input{Message: "Topic Name:"},
		Validate: survey.Required,
	}
	GroupID = &survey.Question{
		Name:     "groupID",
		Prompt:   &survey.Input{Message: "Group ID:"},
		Validate: survey.Required,
	}
)

type InitAnswer struct {
	Module          string
	Dockerfile      bool
	Type            string
	Prometheus      bool
	Database        string
	Framework       string
	Mq              string
	KafkaURL        string
	DBURL           string
	DBName          string
	NameWorker      string
	TopicNameWorker string
	GroupID         string
}

type ServiceAnswer struct {
	Name       string
	Database   string
	NameWorker string
	Module     string
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(serviceCmd)
	rootCmd.AddCommand(repoCmd)
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

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Create Service",
	Run: func(cmd *cobra.Command, args []string) {
		var answers ServiceAnswer

		cobra.CheckErr(startSurveyService(&answers))
		cobra.CheckErr(createServiceProject(answers))
	},
}

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Create Repo",
	Run: func(cmd *cobra.Command, args []string) {
		var answers ServiceAnswer

		cobra.CheckErr(startSurveyRepo(&answers))
		cobra.CheckErr(createRepoProject(answers))
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
		if err := survey.Ask([]*survey.Question{Mq, DBUrl, DBName, KafkaURL, NameWorker, TopicNameWorker, GroupID}, answers); err != nil {
			return err
		}
	}
	return nil
}

func startSurveyService(answers *ServiceAnswer) error {
	if err := survey.Ask([]*survey.Question{DatabaseQuestion, NameWorker}, answers); err != nil {
		return err
	}
	return nil
}

func startSurveyRepo(answers *ServiceAnswer) error {
	if err := survey.Ask([]*survey.Question{DatabaseQuestion, NameWorker}, answers); err != nil {
		return err
	}
	return nil
}

func createServiceProject(answers ServiceAnswer) error {
	absolutePath, err := os.Getwd()
	if err != nil {
		return err
	}

	funcMap := template.FuncMap{
		"ToCamel":      strcase.ToCamel,
		"ToLowerCamel": strcase.ToLowerCamel,
	}

	workerServiceFile, err := os.Create(fmt.Sprintf("%s/internal/%s/%s.service.go", absolutePath, answers.NameWorker, strcase.ToKebab(answers.NameWorker)))
	if err != nil {
		return err
	}

	defer workerServiceFile.Close()

	workerServiceTemplate := template.Must(template.New("workerService").Funcs(funcMap).Parse(tlp.WorkerService))

	if err := workerServiceTemplate.Execute(workerServiceFile, struct {
		MongoDB    bool
		Module     string
		NameWorker string
	}{answers.Database == "MongoDB", absolutePath, answers.NameWorker}); err != nil {
		return err
	}

	return nil
}

func createRepoProject(answers ServiceAnswer) error {
	absolutePath, err := os.Getwd()
	if err != nil {
		return err
	}

	funcMap := template.FuncMap{
		"ToCamel":      strcase.ToCamel,
		"ToLowerCamel": strcase.ToLowerCamel,
		"ToKebab":      strcase.ToKebab,
	}

	workerRepoFile, err := os.Create(fmt.Sprintf("%s/internal/%s/%s.repo.go", absolutePath, answers.NameWorker, strcase.ToKebab(answers.NameWorker)))
	if err != nil {
		return err
	}

	defer workerRepoFile.Close()

	workerRepoTemplate := template.Must(template.New("workerRepo").Funcs(funcMap).Parse(tlp.WorkerRepo))

	if err := workerRepoTemplate.Execute(workerRepoFile, struct {
		MongoDB    bool
		Module     string
		NameWorker string
	}{answers.Database == "MongoDB", answers.Module, answers.NameWorker}); err != nil {
		return err
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

	// write content to env file
	envFile, err := os.OpenFile(fmt.Sprintf("%s/development.env", absolutePath), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer envFile.Close()
	envTemplate := template.Must(template.New("env").Parse(tlp.Env))
	if err := envTemplate.Execute(envFile, struct {
		Kafka           bool
		MongoDB         bool
		DBUrl           string
		DBName          string
		KafkaURL        string
		TopicNameWorker string
	}{answers.Mq == "Kafka", answers.Database == "MongoDB", answers.DBURL, answers.DBName, answers.KafkaURL, answers.TopicNameWorker}); err != nil {
		return err
	}

	// go get mod

	modList := []string{"github.com/rs/zerolog/log", "github.com/joho/godotenv"}
	if answers.Database == "MongoDB" {
		modList = append(modList, "go.mongodb.org/mongo-driver/mongo")
	}
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
		mainTemplate := template.Must(template.New("main").Parse(tlp.MainAPI))
		if err := mainTemplate.Execute(apiFile, struct {
			MongoDB bool
			Gin     bool
		}{answers.Database == "MongoDB", answers.Framework == "Gin"}); err != nil {
			return err
		}
	}

	// create cmd/worker/main.go
	if answers.Type == "Worker" {

		funcMap := template.FuncMap{
			"ToCamel":      strcase.ToCamel,
			"ToLowerCamel": strcase.ToLowerCamel,
		}

		if _, err = os.Stat(fmt.Sprintf("%s/cmd/worker/%s", absolutePath, answers.NameWorker)); os.IsNotExist(err) {
			if err := os.MkdirAll(fmt.Sprintf("%s/cmd/worker/%s", absolutePath, answers.NameWorker), 0751); err != nil {
				return err
			}
		}
		workerFile, err := os.Create(fmt.Sprintf("%s/cmd/worker/%s/main.go", absolutePath, answers.NameWorker))
		if err != nil {
			return err
		}
		defer workerFile.Close()
		mainTemplate := template.Must(template.New("main").Funcs(funcMap).Parse(tlp.MainWorker))
		if err := mainTemplate.Execute(workerFile, struct {
			Kafka           bool
			MongoDB         bool
			Module          string
			KafkaURL        string
			DBUrl           string
			DBName          string
			NameWorker      string
			TopicNameWorker string
			GroupID         string
		}{answers.Mq == "Kafka", answers.Database == "MongoDB", answers.Module, answers.KafkaURL, answers.DBURL, answers.DBName, answers.NameWorker, answers.TopicNameWorker, answers.GroupID}); err != nil {
			return err
		}

		// create {answers.NameWorker}/internal/{answers.NameWorker}.handeler.go
		if _, err = os.Stat(fmt.Sprintf("%s/internal/%s", absolutePath, answers.NameWorker)); os.IsNotExist(err) {
			if err := os.MkdirAll(fmt.Sprintf("%s/internal/%s", absolutePath, answers.NameWorker), 0751); err != nil {
				return err
			}
		}
		workerHandlerFile, err := os.Create(fmt.Sprintf("%s/internal/%s/%s.handler.go", absolutePath, answers.NameWorker, strcase.ToKebab(answers.NameWorker)))
		if err != nil {
			return err
		}

		defer workerHandlerFile.Close()

		workerHandlerTemplate := template.Must(template.New("workerHandler").Funcs(funcMap).Parse(tlp.WorkerHandler))
		if err := workerHandlerTemplate.Execute(workerHandlerFile, struct {
			Kafka      bool
			Module     string
			NameWorker string
		}{answers.Mq == "Kafka", answers.Module, answers.NameWorker}); err != nil {
			return err
		}
	}

	return nil
}
