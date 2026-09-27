# 通用多阶段构建:通过 TARGET 指定要构建的 main 包。
# 用法:
#   docker build -f deploy/Dockerfile --build-arg TARGET=./services/content/cmd -t blog-content .
# 说明:go.mod 要求 go >= 1.26,故基座用 golang:1.26。
# syntax=docker/dockerfile:1

FROM golang:1.26 AS build
WORKDIR /src

# 先拷依赖清单,利用层缓存
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGET=./services/content/cmd
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/app "${TARGET}"

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/app /app/app
# 服务配置文件(只读;也可由外部挂载覆盖)
COPY --from=build /src/services /app/services
USER nonroot:nonroot
ENTRYPOINT ["/app/app"]
