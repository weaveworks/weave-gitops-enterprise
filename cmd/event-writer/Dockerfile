FROM golang:1.14 as builder
ENV EWPATH=cmd/event-writer

WORKDIR /workspace/$EWPATH

# copy in other module dependencies
COPY /common/ /workspace/common/

COPY $EWPATH/go.mod go.mod
COPY $EWPATH/go.sum go.sum

# cache modules
RUN go mod download

# copy source code
COPY $EWPATH/main.go main.go
COPY $EWPATH/converter/ converter/
COPY $EWPATH/database/ database/
COPY $EWPATH/queue/ queue/
COPY $EWPATH/run/ run/
COPY $EWPATH/subscribe/ subscribe/
COPY $EWPATH/messages/ messages/
COPY $EWPATH/liveness/ liveness/


# build
RUN CGO_ENABLED=1 go build -ldflags '-linkmode external -w -extldflags "-static"' -o event-writer main.go

FROM alpine:3.12
ENV EWPATH=cmd/event-writer

COPY --from=builder /workspace/$EWPATH/event-writer /event-writer

COPY $EWPATH/start.sh /
RUN apk add --no-cache tini

ENTRYPOINT [ "/sbin/tini", "--", "/start.sh"]
