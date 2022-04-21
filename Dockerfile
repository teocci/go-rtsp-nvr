FROM golang:1.17-alpine3.14
RUN apk add --no-cache make docker-cli git gcc musl-dev pkgconfig ffmpeg-dev
WORKDIR /s
COPY go.mod go.sum ./
RUN go mod download
COPY src/rtsp-server ./