FROM alpine

RUN apk add --no-cache ca-certificates tini

RUN addgroup -S clusters-service && adduser -S clusters-service -G clusters-service

COPY ./cmd/clusters-service/bin /app
