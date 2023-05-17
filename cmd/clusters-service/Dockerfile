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

# Use distroless as minimal base image
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot

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
USER 65532:65532

ENTRYPOINT ["/clusters-service"]