#FROM ubuntu:latest
#LABEL authors="CodeMachine"
#
#ENTRYPOINT ["top", "-b"]

# 第一阶段：构建
FROM golang:1.24.3 AS builder

# 设置环境变量
ENV GO111MODULE=on
ENV GOPROXY=http://goproxy.cn,direct

# 设置工作目录
WORKDIR /build

# 复制项目文件
COPY . .

# 构建应用
WORKDIR /build/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /output/api .

# 第二阶段：运行
FROM alpine:latest

# 创建配置目录
RUN mkdir -p /api/etc

# 复制配置文件和可执行文件
COPY --from=builder /build/api/etc/chat-docker.yaml /api/etc/chat.yaml
COPY --from=builder /output/api /api/

WORKDIR /api
CMD ["./api"]