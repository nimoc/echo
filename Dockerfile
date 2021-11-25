FROM golang:1.15.6-alpine


# 时区配置
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' > /etc/timezone
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add tzdata


# go模块代理
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct

WORKDIR /app

COPY main.go main.go
COPY go.mod go.mod

RUN go build -o server

EXPOSE 8080

ENTRYPOINT ["/app/server"]
