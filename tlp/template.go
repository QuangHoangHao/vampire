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
)
