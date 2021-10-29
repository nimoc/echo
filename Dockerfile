FROM golang:1.15.6-alpine


# # 时区配置
# 有时候构建会安装 tzdata 特别慢,为了避免学习k8s的时候被这个慢打断就先注释
# RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
# RUN echo 'Asia/Shanghai' > /etc/timezone
# RUN apk add tzdata


# go模块代理
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct

WORKDIR /app

COPY main.go main.go
COPY go.mod go.mod

RUN go build -o server

EXPOSE 8080

ENTRYPOINT ["/app/server"]
