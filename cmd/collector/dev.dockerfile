FROM alpine

RUN apk add --no-cache ca-certificates tini

RUN addgroup -S collector && adduser -S collector -G collector

COPY ./cmd/collector/bin /app

ENTRYPOINT [ "/sbin/tini", "--", "collector" ]
