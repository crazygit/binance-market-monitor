FROM golang:1.18.1-alpine as builder

ENV GO111MODULE=on \
    CGO_ENABLED=0

ENV APP_NAME=binance-market-monitor
WORKDIR /build

# no such package in aarch64
# RUN apk add --no-cache upx

# Copy and download dependency using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -o binance-market-monitor

#FROM alpine:3
FROM scratch
WORKDIR /app
# 修复使用scratch时报没有证书的错误
# copy the ca-certificate.crt from the build stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /build/binance-market-monitor .


ENV TELEGRAM_API_TOKEN \
    TELEGRAM_CHANNEL_USERNAME \
    ENVIRONMENT=production \
    QUOTE_ASSET=USDT \
    LOWEST_PRICE_FILTER=1.0 \
    PRICE_CHANGE_PERCENT_THRESHOLD=5.0 \
    PRICE_CHANGE_THRESHOLD=0.5 \

CMD ["/app/binance-market-monitor"]
