FROM golang:1.14 as builder

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum

# cache modules
RUN go mod download

# copy source code
COPY main.go main.go
COPY converter/ converter/
COPY database/ database/
COPY run/ run/
COPY subscribe/ subscribe/
COPY messages/ messages/

# build
RUN CGO_ENABLED=1 go build -ldflags '-linkmode external -w -extldflags "-static"' -o event-writer main.go

FROM alpine:3.12

COPY --from=builder /workspace/event-writer /event-writer

COPY start.sh /
RUN apk add --no-cache tini

ENTRYPOINT [ "/sbin/tini", "--", "/start.sh"]
