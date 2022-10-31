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
)
