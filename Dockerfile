FROM golang:1.15.6-alpine

# go模块代理
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct

WORKDIR /app

COPY main.go main.go
COPY go.mod go.mod

RUN go build -o server

EXPOSE 8080

ENTRYPOINT ["/app/server"]
