FROM golang:alpine AS builder

# 为镜像设置必要的环境变量
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# 移动到工作目录：/build
WORKDIR /build

# 复制项目中的go.mod和go.sum文件并下载依赖信息
COPY go.mod .
COPY go.sum .
RUN go mod download

# 将代码复制到容器中
COPY . .

# 将我们的代码编译成二进制可执行文件shit
RUN go build -o checkin_app .

#############
# 创建一个小镜像
#############
FROM alpine:3.18

WORKDIR /app

# 从builder镜像中把/build/checkin拷贝到当前目录
COPY --from=builder /build/checkin_app /app/
COPY ./wait-for.sh /app/
COPY ./conf/ /app/conf/

RUN set -eux; \
    apk add --no-cache \
        netcat-openbsd \
        bash; \
    chmod +x /app/wait-for.sh /app/checkin_app; \
    chmod -R 755 /app/conf

# 声明服务接口
EXPOSE 8084

# 需要运行的命令
#ENTRYPOINT ["/shit_app", "conf/config.yaml"]
