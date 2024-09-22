# builder
FROM golang:alpine AS builder

COPY go.mod go.sum ./

RUN go mod download -x

RUN mkdir -p /result && \
    apk add --no-cache openssl && \
    openssl req -x509 -newkey rsa:4096 -sha256 -days 365 -nodes -keyout /result/example.key -out /result/example.crt -subj "/CN=localhost"

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -ldflags "-s -w" -a -installsuffix cgo -o /result/https-proxy ./main.go

# result image
FROM alpine

RUN apk add --no-cache ca-certificates && \
    addgroup -g 1001 proxy && \
    adduser -h /proxy -u 1001 -G proxy -s /bin/sh -D proxy

WORKDIR /proxy

COPY --from=builder --chown=proxy:proxy /result /proxy

USER proxy

CMD ["/proxy/https-proxy"]