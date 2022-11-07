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
	MainAPI string = `package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	{{ if .MongoDB}} 
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	{{end}}
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
	{{ if .MongoDB}} 
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URL")))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Error().Err(err).Msg("")
	}
	log.Info().Msg("Connected to MongoDB")
	
	client.Database(os.Getenv("MONGO_DB_NAME"))
	{{end}}

	{{if .Gin}}
	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	router.Run(":3000")
	log.Info().Msg("server is running at : 3000")
	{{end}}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")
}`
	MainWorker string = `package main

import (
	"context"
	"os"
	"os/signal"
	{{if .Kafka}} 
	"{{ .Module }}/pkg/kafka_sesu"
	{{end}}
	"syscall"
	"time"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	{{if .MongoDB}} 
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	{{end}}
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	{{if .MongoDB}} 
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("{{ .DBUrl }}")))
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Error().Err(err).Msg("")
	}
	log.Info().Msg("Connected to MongoDB")
	{{end}}

	// database := client.Database(os.Getenv("{{ .DBName }}"))

	// {{ .NameWorker | ToLowerCamel }}Repo := {{ .NameWorker }}.New{{ .NameWorker | ToCamel }}Repository(database)
	// {{ .NameWorker | ToLowerCamel }}Service := {{ .NameWorker }}.New{{ .NameWorker | ToCamel }}Service({{ .NameWorker | ToLowerCamel }}Repo)
	// {{ .NameWorker | ToLowerCamel }}Handler := {{ .NameWorker }}.New{{ .NameWorker | ToCamel }}Handler({{ .NameWorker | ToLowerCamel }}Service)

	{{ .NameWorker | ToLowerCamel }}KafkaHandler := make(map[string]kafka_sesu.CallHandler)
	// {{ .NameWorker | ToLowerCamel }}KafkaHandler[os.Getenv("{{ .TopicNameWorker }}")] = {{ .NameWorker | ToLowerCamel }}Handler.

	{{ .NameWorker | ToLowerCamel }}Consumer := kafka_sesu.NewConsumerKafka(kafka_sesu.ConsumerConfig{
		Brokers:           []string{os.Getenv("{{ .KafkaURL }}")},
		GroupID:           "{{ .GroupID }}",
		MinBytes:          10e2,
		MaxBytes:          10e3,
		HeartbeatInterval: time.Second * 20,
		SessionTimeout:    time.Second * 60,
		MaxWait:           time.Second * 1,
		Handler:           {{ .NameWorker | ToLowerCamel }}KafkaHandler,
	})
	go {{ .NameWorker | ToLowerCamel }}Consumer.Consume()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		log.Info().Msg("app - Run - signal: " + s.String())
	case <-{{ .NameWorker | ToLowerCamel }}Consumer.Notify():
		log.Info().Msg("{{ .NameWorker | ToLowerCamel }}Consumer crash!")
	}
	defer {{ .NameWorker | ToLowerCamel }}Consumer.Shutdown()
}`

	WorkerHandler string = `package {{ .NameWorker }}

import (
	// "encoding/json"
	{{if .Kafka}} 
	// "{{ .Module }}/pkg/kafka_sesu"
	{{end}}
	// "{{ .Module }}/internal/common"
	// "github.com/rs/zerolog/log"
)

type {{ .NameWorker | ToCamel }}Hander interface {

}

type {{ .NameWorker | ToLowerCamel }}Handler struct {
	{{ .NameWorker | ToLowerCamel }}Service {{ .NameWorker | ToCamel }}Service
}

func New{{ .NameWorker | ToCamel }}Handler({{ .NameWorker | ToLowerCamel }}Service {{ .NameWorker | ToCamel }}Service) {{ .NameWorker | ToCamel }}Hander {
	return &{{ .NameWorker | ToLowerCamel }}Handler{ {{ .NameWorker | ToLowerCamel }}Service: {{ .NameWorker | ToLowerCamel }}Service}
}
`

	WorkerService string = `package {{ .NameWorker }}

import (
	// "encoding/json"
	// "errors"
	// "{{ .Module }}/internal/common"
	// "strconv"
	// "time"
	// "github.com/rs/zerolog/log"
	{{if .MongoDB}} 
	// "context"
	// "go.mongodb.org/mongo-driver/mongo"
	{{end}}
)

type {{ .NameWorker | ToCamel }}Service interface {

}

type {{ .NameWorker | ToLowerCamel }}Service struct {
	repository {{ .NameWorker | ToCamel }}Repository
}

func New{{ .NameWorker | ToCamel }}Service(repository {{ .NameWorker | ToCamel }}Repository) {{ .NameWorker | ToCamel }}Service {
	return &{{ .NameWorker | ToLowerCamel }}Service{repository: repository}
}
`

	WorkerRepo string = `package {{ .NameWorker }}
import (
	// "{{ .Module }}/internal/common"
	// "time"
	// "github.com/rs/zerolog/log"
	{{if .MongoDB}} 
	// "context"
	// "go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	// "go.mongodb.org/mongo-driver/mongo/options"
	{{end}}
)

type {{ .NameWorker | ToCamel }}Repository interface {

}

type {{ .NameWorker | ToLowerCamel }}Repository struct {
	{{if .MongoDB}} 
	database      *mongo.Database
	collection    *mongo.Collection
	{{end}}
}

func New{{ .NameWorker | ToCamel }}Repository({{if .MongoDB}}database *mongo.Database{{end}}) {{ .NameWorker | ToCamel }}Repository {
	return &{{ .NameWorker | ToLowerCamel }}Repository{
		{{if .MongoDB}}
		database:      database,
		collection:    database.Collection("{{ .NameWorker | ToKebab }}"),
		{{end}}
	}
}
`

	Env string = `APP_ENV=development
port=3000
{{if .MongoDB}}
{{ .DBUrl }}=
{{ .DBName }}=
{{end}}
{{ if .Kafka -}}
{{ .KafkaURL }}=
{{ .TopicNameWorker }}=
{{- end }}
`
)
