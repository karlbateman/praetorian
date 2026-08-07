FROM golang:1.26.5-alpine3.24 AS builder
WORKDIR /build

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

COPY . .
RUN go build -ldflags="-w -s" -o /main ./cmd

RUN apk update && apk add --no-cache upx
RUN upx --best /main

FROM scratch
COPY --from=builder /main /main
ENTRYPOINT [ "/main" ]
