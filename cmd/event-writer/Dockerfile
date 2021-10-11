FROM golang:1.17 as builder

ENV EW_PATH=cmd/event-writer

WORKDIR /workspace/$EW_PATH

# copy in other module dependencies
COPY /common/ /workspace/common/

# copy modules manifests
COPY $EW_PATH/go.mod go.mod
COPY $EW_PATH/go.sum go.sum

ENV GOPRIVATE=github.com/weaveworks/weave-gitops-enterprise-credentials
ENV GITHUB_BUILD_USERNAME=wge-build-bot
ARG GITHUB_BUILD_TOKEN
RUN git config --global url."https://${GITHUB_BUILD_USERNAME}:${GITHUB_BUILD_TOKEN}@github.com".insteadOf "https://github.com"

# cache modules
RUN go mod download

# copy source code
COPY $EW_PATH/main.go main.go
COPY $EW_PATH/converter/ converter/
COPY $EW_PATH/database/ database/
COPY $EW_PATH/queue/ queue/
COPY $EW_PATH/run/ run/
COPY $EW_PATH/subscribe/ subscribe/
COPY $EW_PATH/liveness/ liveness/

# build
RUN CGO_ENABLED=1 go build -ldflags '-linkmode external -w -extldflags "-static"' -o event-writer main.go

FROM alpine:3.14
ENV EW_PATH=cmd/event-writer

COPY --from=builder /workspace/$EW_PATH/event-writer /event-writer

COPY $EW_PATH/start.sh /
RUN apk add --no-cache tini

ENTRYPOINT [ "/sbin/tini", "--", "/start.sh"]
