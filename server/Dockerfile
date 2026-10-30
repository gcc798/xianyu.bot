# syntax=docker/dockerfile:1
FROM golang:1.26.10-alpine AS builder

WORKDIR /go/src/github.com/flipped-aurora/gin-vue-admin/server
RUN go env -w GO111MODULE=on \
    && go env -w GOPROXY=https://goproxy.cn,direct \
    && go env -w CGO_ENABLED=0

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o server . \
    && go build -o initdb ./cmd/initdb

FROM alpine:latest

LABEL MAINTAINER="SliverHorn@sliver_horn@qq.com"
# 设置时区
ENV TZ=Asia/Shanghai
RUN apk update && apk add --no-cache tzdata openntpd \
    && ln -sf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /go/src/github.com/flipped-aurora/gin-vue-admin/server

COPY --from=0 /go/src/github.com/flipped-aurora/gin-vue-admin/server/server ./
COPY --from=0 /go/src/github.com/flipped-aurora/gin-vue-admin/server/initdb ./
COPY --from=0 /go/src/github.com/flipped-aurora/gin-vue-admin/server/resource ./resource/
COPY --from=0 /go/src/github.com/flipped-aurora/gin-vue-admin/server/config.docker.yaml ./

# SQLite 持久化示例：docker run -d -v /宿主机路径/xianyu_assistant.db:/go/src/github.com/flipped-aurora/gin-vue-admin/server/xianyu_assistant.db -p 8888:8888 --name gva-server-v1 gva-server:1.0
# VOLUME ["/go/src/github.com/flipped-aurora/gin-vue-admin/server"]

EXPOSE 8888
ENTRYPOINT ./server -c config.docker.yaml
