# ---- 阶段一: Builder ----
FROM  golang:1.25-alpine3.22 AS builder


WORKDIR /src

#缓存了 apk 的包索引和下载的包。
RUN --mount=type=cache,target=/var/cache/apk \
    sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && \
    apk add --no-cache tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone


# 只复制 go.mod 和 go.sum，这样只有在它们改变时才重新下载依赖
COPY go.mod go.sum ./


RUN go env -w GOPROXY=https://goproxy.cn,direct


#缓存了 Go 模块，这是最重要的优化。
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download


COPY . .


# target=/root/.cache/go-build 缓存了 Go 编译器的中间产物，加快了重新编译的速度。
ARG TARGETOS TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o /app/ro3_api2 .


# ---- 阶段2: Final Image ----
FROM alpine:3.22

# 优化最终镜像的 APK 安装，同样使用缓存
RUN --mount=type=cache,target=/var/cache/apk \
    sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && \
    apk add --no-cache ca-certificates tzdata curl bind-tools && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app
COPY --from=builder /app/ro3_api2 /app/ro3_api2
COPY ./etc /app/etc

# 使用非 root 用户运行（安全优化）
#RUN adduser -D -u 10001 appuser
#USER appuser

CMD ["./ro3_api2", "-f", "etc/ro3-api.yaml"]
