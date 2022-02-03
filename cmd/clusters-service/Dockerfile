# --
# -- node build env
FROM node:14.19 as node_with_modules

WORKDIR /workspace/

# Copy package.json stuff and install modules
COPY ui-cra/package.json ui-cra/yarn.lock ui-cra/

RUN echo "@weaveworks:registry=https://npm.pkg.github.com" >> ui-cra/.npmrc
RUN echo --mount=type=secret,id=GITHUB_BUILD_TOKEN "//npm.pkg.github.com/:_authToken=$(cat /run/secrets/GITHUB_BUILD_TOKEN)" >> ui-cra/.npmrc

RUN yarn config set network-timeout 300000 && \
  cd ui-cra && \
  yarn install --frozen-lockfile

# --
# -- golang build env
FROM golang:1.17.1 as go_with_modules

WORKDIR /workspace/

ENV GITHUB_BUILD_USERNAME=wge-build-bot
RUN --mount=type=secret,id=GITHUB_BUILD_TOKEN \
  git config --global url."https://${GITHUB_BUILD_USERNAME}:$(cat /run/secrets/GITHUB_BUILD_TOKEN)@github.com".insteadOf "https://github.com"

# copy modules manifests
COPY go.mod /workspace/go.mod
COPY go.sum /workspace/go.sum

# copy in other module dependencies
COPY common/ /workspace/common/

# cache modules
RUN go mod download

# --
# -- build ui
FROM node_with_modules as node_builder

WORKDIR /workspace/ui-cra

ARG version

# Copy source and supporting files and build
COPY ui-cra/tsconfig.json .
COPY ui-cra/src src
COPY ui-cra/public public
RUN REACT_APP_VERSION=$version yarn build

# --
# -- build go
FROM go_with_modules as go_builder

WORKDIR /workspace/cmd/clusters-service

# common code
COPY pkg/ /workspace/pkg/

# copy source code
COPY cmd/clusters-service/main.go main.go
COPY cmd/clusters-service/api/ api/
COPY cmd/clusters-service/app/ app/
COPY cmd/clusters-service/pkg/ pkg/
COPY --from=node_builder /workspace/ui-cra/build app/dist

ARG version
ARG image_tag
# build
RUN CGO_ENABLED=1 go build -ldflags "-X 'github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/version.Version=$version -X github.com/weaveworks/weave-gitops-enterprise/pkg/version.ImageTag=$image_tag' -linkmode external -w -extldflags \"-static\"" -tags netgo -a -o clusters-service main.go

# --
# -- runtime!
FROM docker.io/governmentpaas/git-ssh

ARG now
ARG revision
ARG version
LABEL maintainer="Weaveworks <help@weave.works>" \
      org.opencontainers.image.created="${now}" \
      org.opencontainers.image.description="The Weave GitOps Enterprise Clusters Service" \
      org.opencontainers.image.documentation="https://docs.gitops.weave.works/" \
      org.opencontainers.image.revision="${revision}" \
      org.opencontainers.image.title="weave-gitops-enterprise-clusters-service" \
      org.opencontainers.image.url="https://docs.gitops.weave.works/" \
      org.opencontainers.image.vendor="Weaveworks" \
      org.opencontainers.image.version="${version}"


# # Add git provider hostnames to known hosts file so we can use
# # StrickHostKeyChecking with git+ssh
COPY ./tools/known_hosts.sh /tmp/known_hosts.sh
RUN bash /tmp/known_hosts.sh /etc/ssh/ssh_known_hosts && rm /tmp/known_hosts.sh

RUN apk add --no-cache ca-certificates tini

COPY --from=go_builder /workspace/cmd/clusters-service/clusters-service /usr/local/bin/

RUN addgroup -S clusters-service && adduser -S clusters-service -G clusters-service

#
# FIXME: it would be nice not run as root but k8s mounts the sqlite host volume as root:root
# and then the clusters-service user can't write to it.
#
# USER clusters-service
# RUN mkdir -p /home/clusters-service/.ssh

ENTRYPOINT [ "/sbin/tini", "--", "clusters-service" ]
