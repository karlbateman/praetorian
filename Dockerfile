FROM golang:1.24.1-alpine3.21 AS builder
WORKDIR /build

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

COPY . .
RUN go build -ldflags="-w -s" -o /main ./cmd

RUN apk update && apk add --no-cache upx
RUN upx --brute /main

FROM scratch
COPY --from=builder /main /main
ENTRYPOINT [ "/main" ]