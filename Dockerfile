FROM golang:1.25-bookworm

# 开启 Go 模块代理（解决国内下载依赖失败）
ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=0

WORKDIR /app

COPY repo/ .

# 构建项目
RUN go mod tidy
RUN go build -o app .

CMD ["./app", "serve"]