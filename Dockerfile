FROM golang:1.23.0-alpine3.20 AS builder

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
  -o /bin/renovate-controller . \
  && strip /bin/renovate-controller \
  && upx -q -9 /bin/renovate-controller

FROM alpine:3.20

COPY --from=builder /bin/renovate-controller /usr/local/renovate-controller

RUN addgroup -S gouser && adduser -S -G gouser -s /sbin/nologin gouser

USER gouser

ENTRYPOINT ["/usr/local/renovate-controller"]
