# 前端构建阶段
FROM node:18 AS frontend-builder
WORKDIR /app
# 安装 patch 工具
RUN apt-get update && apt-get install -y patch
# 复制 excalidraw 子模块
COPY excalidraw/ ./excalidraw/
# 复制补丁文件
COPY frontend.patch ./
# 应用前端补丁
RUN cd excalidraw && patch -p1 < ../frontend.patch
# 构建前端
RUN cd excalidraw && npm install && cd excalidraw-app && npm run build:app:docker

# 后端构建阶段
FROM golang:alpine AS backend-builder
RUN apk update && apk add --no-cache git
WORKDIR /app
# 复制 Go 模块文件
COPY go.mod go.sum ./
RUN GOPROXY=direct go mod download
# 复制源代码
COPY . .
# 复制前端构建文件到正确位置，以便 Go embed 可以找到
COPY --from=frontend-builder /app/excalidraw/excalidraw-app/build ./frontend/
# 构建 Go 应用
RUN GOPROXY=direct CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 最终运行镜像
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
# 复制后端二进制文件（已包含嵌入的前端文件）
COPY --from=backend-builder /app/main .
# 暴露端口
EXPOSE 3002
CMD ["./main"]
