FROM golang:1.14-alpine as builder

LABEL MAINTAINER nattapong@chochoe.net
LABEL APP pg-ping

WORKDIR /app
COPY . /app

ENV CGO_ENABLED=0
RUN apk update && apk upgrade \
    && apk add --no-cache ca-certificates git \
    && rm -rf /var/cache/apk/*
RUN go mod download
RUN go build -a --ldflags "-s -w" -o /release/pg-ping

FROM alpine:3.11

RUN apk update && apk upgrade \
  && apk add ca-certificates \
  && rm -rf /var/cache/apk/*

WORKDIR /app
ENV PATH='$PATH:/app'
COPY --from=builder /release/* ./

CMD ["pg-ping"]
