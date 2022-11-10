package tlp

var (
	ReadmeTemplate string = `# Golang App`
	Gitignore      string = `/app/log/*
/app/.env
.vscode
.env
.idea
vendor
/tmp`
	Dockerfile string = `FROM golang:1.19.0-alpine3.15 AS builder

WORKDIR /app

COPY go.mod go.sum ./
	
RUN go mod download
	
COPY . /app

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -o /app/bin/ /app/cmd/...

FROM alpine:3.15.6

WORKDIR /app

COPY --from=builder /app/bin /app/bin
COPY --from=builder /app/*.env /app`
	API string = `package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})

	fileEnv := "development.env"
	if os.Getenv("APP_ENV") == "production" {
		fileEnv = "production.env"
	}
	if err := godotenv.Load(fileEnv); err != nil {
		log.Error().Err(err).Msg("parse env failed")
	}

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	router.Run(":3000")
	log.Info().Msg("server is running at : 3000")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")
}`
	Worker string = `package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)
	
func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})

	fileEnv := "development.env"
	if os.Getenv("APP_ENV") == "production" {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		fileEnv = "production.env"
	}
	if err := godotenv.Load(fileEnv); err != nil {
		log.Error().Err(err).Msg("parse env failed")
		return
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		log.Info().Msg("app - Run - signal: " + s.String())
	}
}`

	WorkerHandler string = `package {{ .WorkerName }}

import (
	// "encoding/json"
	// "github.com/rs/zerolog/log"
)

type {{ .WorkerName | ToCamel }}Hander interface {
}

type {{ .WorkerName | ToLowerCamel }}Handler struct {
}

func New{{ .WorkerName | ToCamel }}Handler() {{ .WorkerName | ToCamel }}Hander {
	return &{{ .WorkerName | ToLowerCamel }}Handler{}
}
`

	Controller string = `package {{ .Name }}

import (
	// "encoding/json"
	// "net/http"
	// "strconv"
	// "time"

	// "github.com/gin-gonic/gin"
	// "github.com/rs/zerolog/log"
)

type {{ .Name | ToCamel }}Controller interface {
}

type {{ .Name | ToLowerCamel }}Controller struct {
	{{ .Name | ToLowerCamel }}Service {{ .Name | ToCamel }}Service
}

func New{{ .Name | ToCamel }}Controller({{ .Name | ToLowerCamel }}Service {{ .Name | ToCamel }}Service) {{ .Name | ToCamel }}Controller {
	return &{{ .Name | ToLowerCamel }}Controller{ {{ .Name | ToLowerCamel }}Service: {{ .Name | ToLowerCamel }}Service }
}
`

	Service string = `package {{ .Name }}

import (
	// "encoding/json"
	// "errors"
	// "strconv"
	// "time"
	// "github.com/rs/zerolog/log"
)

type {{ .Name | ToCamel }}Service interface {
}

type {{ .Name | ToLowerCamel }}Service struct {
	{{ .Name | ToLowerCamel }}Repository {{ .Name | ToCamel }}Repository
}

func New{{ .Name | ToCamel }}Service({{ .Name | ToLowerCamel }}Repository {{ .Name | ToCamel }}Repository) {{ .Name | ToCamel }}Service {
	return &{{ .Name | ToLowerCamel }}Service{ {{ .Name | ToLowerCamel }}Repository: {{ .Name | ToLowerCamel }}Repository }
}
`

	Repository string = `package {{ .Name }}

import (
	// "time"
	// "github.com/rs/zerolog/log"
	{{if .MongoDB}} 
	// "context"
	// "go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/bson/primitive"
	// "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo"
	{{end}}
)

type {{ .Name | ToCamel }}Repository interface {
}

type {{ .Name | ToLowerCamel }}Repository struct {
	{{if .MongoDB}}database      *mongo.Database
	collection    *mongo.Collection{{end}}
}

func New{{ .Name | ToCamel }}Repository({{if .MongoDB}}database *mongo.Database{{end}}) {{ .Name | ToCamel }}Repository {
	return &{{ .Name | ToLowerCamel }}Repository{
		{{if .MongoDB}}database:      database,
		collection:    database.Collection("{{ .Name | ToKebab }}"),{{end}}
	}
}
`
	Entity string = `package {{ .Name }}
{{if .MongoDB}}
import "go.mongodb.org/mongo-driver/bson/primitive"
{{end}}
type {{ .Name | ToCamel }} struct {
	{{if .MongoDB}}ID primitive.ObjectID ` + "`" + `json:"id" bson:"_id"` + "`" + `{{end}}
}
`

	Env string = `APP_ENV=development
port=3000
`
)
