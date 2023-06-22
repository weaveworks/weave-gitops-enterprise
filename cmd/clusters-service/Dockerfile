# syntax = docker/dockerfile:1.3

FROM golang:1.20 AS build

WORKDIR /app
COPY Makefile /app/
COPY go.* /app/
COPY common/go.* /app/common/

ENV GITHUB_BUILD_USERNAME=wge-build-bot
ARG GITHUB_BUILD_TOKEN
RUN git config --global url."https://${GITHUB_BUILD_USERNAME}:${GITHUB_BUILD_TOKEN}@github.com".insteadOf "https://github.com"
RUN go mod download

COPY cmd/clusters-service /app/cmd/clusters-service
COPY common /app/common
COPY pkg /app/pkg

ARG version
ARG image_tag
RUN --mount=type=cache,target=/root/.cache/go-build \
      make cmd/clusters-service/clusters-service WEAVE_GITOPS_VERSION=${version} IMAGE_TAG=${image_tag}

FROM alpine:3.18

ARG now
ARG revision
LABEL maintainer="Weaveworks <help@weave.works>" \
      org.opencontainers.image.created="${now}" \
      org.opencontainers.image.description="The Weave GitOps Enterprise Clusters Service" \
      org.opencontainers.image.documentation="https://docs.gitops.weave.works/" \
      org.opencontainers.image.revision="${revision}" \
      org.opencontainers.image.title="weave-gitops-enterprise-clusters-service" \
      org.opencontainers.image.url="https://docs.gitops.weave.works/" \
      org.opencontainers.image.vendor="Weaveworks" \
      org.opencontainers.image.version="${version}"

WORKDIR /
COPY --from=build /app/cmd/clusters-service/clusters-service .


RUN apk --no-cache add ca-certificates tini \
    # gpg is needed for sops encryption and decryption
    gnupg \
    # upgrade openssl to get latest security patches
    && apk upgrade libssl3 libcrypto3 \
    && update-ca-certificates

ENV GNUPGHOME=/tmp/

# nobody user
USER 65534:65534

ENTRYPOINT [ "/sbin/tini", "--", "/clusters-service"]
