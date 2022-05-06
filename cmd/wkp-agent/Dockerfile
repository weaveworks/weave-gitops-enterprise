FROM golang:1.17 AS build

WORKDIR /app
COPY Makefile /app/
COPY go.* /app/
COPY common/go.* /app/common/

ENV GITHUB_BUILD_USERNAME=wge-build-bot
ARG GITHUB_BUILD_TOKEN
RUN git config --global url."https://${GITHUB_BUILD_USERNAME}:${GITHUB_BUILD_TOKEN}@github.com".insteadOf "https://github.com"
RUN go mod download

COPY cmd/wkp-agent /app/cmd/wkp-agent
COPY common /app/common
COPY pkg /app/pkg

ARG version
ARG image_tag
RUN make cmd/wkp-agent/wkp-agent WEAVE_GITOPS_VERSION=${version} IMAGE_TAG=${image_tag}

FROM docker.io/governmentpaas/git-ssh

ARG now
ARG revision
ARG version
LABEL maintainer="Weaveworks <help@weave.works>" \
      org.opencontainers.image.created="${now}" \
      org.opencontainers.image.description="The Weave GitOps Enterprise WKP Agent" \
      org.opencontainers.image.documentation="https://docs.gitops.weave.works/" \
      org.opencontainers.image.revision="${revision}" \
      org.opencontainers.image.title="wkp-agent" \
      org.opencontainers.image.url="https://docs.gitops.weave.works/" \
      org.opencontainers.image.vendor="Weaveworks" \
      org.opencontainers.image.version="${version}"

COPY --from=build /app/cmd/wkp-agent/wkp-agent /

ENTRYPOINT ["/wkp-agent"]

CMD ["--watch"]
