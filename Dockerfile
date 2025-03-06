FROM golang:1.24.0-alpine3.21 AS builder

ENV GO111MODULE=on \
  CGO_ENABLED=1 \
  GOOS=linux \
  GOARCH=amd64

RUN apk update && apk upgrade
RUN apk add upx \
  gcc \
  musl-dev

WORKDIR /src
COPY . .

RUN go build \
  -ldflags "-s -w -extldflags '-static'" \
  -o /bin/renovate-controller .\cmd\controller\main.go \
  && strip /bin/renovate-controller \
  && upx -q -9 /bin/renovate-controller

FROM alpine:3.21

COPY --from=builder /bin/renovate-controller /usr/local/renovate-controller

RUN addgroup -S gouser && adduser -S -G gouser -s /sbin/nologin gouser

RUN mkdir /data && chown gouser:gouser /data
VOLUME /data

USER gouser

ENTRYPOINT ["/usr/local/renovate-controller"]
