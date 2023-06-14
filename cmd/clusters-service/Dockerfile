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
# Use `base` variant as we compile with CGO_ENABLED=1 for sqlite3 and it includes glibc
FROM gcr.io/distroless/base:nonroot

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

# gpg is needed for sops encryption and decryption
COPY --from=build /usr/bin/gpg /usr/bin/gpg
# required libraries to run gpg
COPY --from=build /lib/x86_64-linux-gnu/libz.so.1 /lib/x86_64-linux-gnu/libz.so.1
COPY --from=build /lib/x86_64-linux-gnu/libbz2.so.1.0 /lib/x86_64-linux-gnu/libbz2.so.1.0
COPY --from=build /usr/lib/x86_64-linux-gnu/libsqlite3.so.0 /usr/lib/x86_64-linux-gnu/libsqlite3.so.0
COPY --from=build /usr/lib/x86_64-linux-gnu/libgcrypt.so.20 /usr/lib/x86_64-linux-gnu/libgcrypt.so.20
COPY --from=build /lib/x86_64-linux-gnu/libreadline.so.8 /lib/x86_64-linux-gnu/libreadline.so.8
COPY --from=build /usr/lib/x86_64-linux-gnu/libassuan.so.0 /usr/lib/x86_64-linux-gnu/libassuan.so.0
COPY --from=build /lib/x86_64-linux-gnu/libgpg-error.so.0 /lib/x86_64-linux-gnu/libgpg-error.so.0
COPY --from=build /lib/x86_64-linux-gnu/libc.so.6 /lib/x86_64-linux-gnu/libc.so.6
COPY --from=build /lib/x86_64-linux-gnu/libm.so.6 /lib/x86_64-linux-gnu/libm.so.6
COPY --from=build /lib/x86_64-linux-gnu/libpthread.so.0 /lib/x86_64-linux-gnu/libpthread.so.0
COPY --from=build /lib/x86_64-linux-gnu/libdl.so.2 /lib/x86_64-linux-gnu/libdl.so.2
COPY --from=build /lib/x86_64-linux-gnu/libtinfo.so.6 /lib/x86_64-linux-gnu/libtinfo.so.6
COPY --from=build /lib64/ld-linux-x86-64.so.2 /lib64/ld-linux-x86-64.so.2

USER 65532:65532

ENTRYPOINT ["/clusters-service"]
