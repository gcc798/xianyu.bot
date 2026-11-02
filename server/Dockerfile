# syntax=docker/dockerfile:1
FROM golang:1.26.10-alpine AS builder

WORKDIR /go/src/github.com/flipped-aurora/gin-vue-admin/server
RUN go env -w GO111MODULE=on \
    && go env -w CGO_ENABLED=0

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o server .

FROM alpine:latest

LABEL org.opencontainers.image.title="Xianyu Harness Agent"
# 设置时区
ENV TZ=Asia/Shanghai
RUN apk update && apk add --no-cache tzdata openntpd \
    && ln -sf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /go/src/github.com/flipped-aurora/gin-vue-admin/server

COPY --from=0 /go/src/github.com/flipped-aurora/gin-vue-admin/server/server ./
COPY --from=0 /go/src/github.com/flipped-aurora/gin-vue-admin/server/config.docker.yaml ./

EXPOSE 8888
ENTRYPOINT ["./server", "-c", "config.docker.yaml"]
