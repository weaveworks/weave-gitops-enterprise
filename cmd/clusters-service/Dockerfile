FROM golang:1.17.1 as builder

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

ENV CS_PATH=cmd/clusters-service

WORKDIR /workspace/$CS_PATH

ENV GITHUB_BUILD_USERNAME=wge-build-bot
ARG GITHUB_BUILD_TOKEN
RUN git config --global url."https://${GITHUB_BUILD_USERNAME}:${GITHUB_BUILD_TOKEN}@github.com".insteadOf "https://github.com"

# copy modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# copy in other module dependencies
COPY common/ /workspace/common/

# cache modules
RUN go mod download

# copy source code
COPY $CS_PATH/main.go main.go
COPY $CS_PATH/api/ api/
COPY $CS_PATH/app/ app/
COPY $CS_PATH/pkg/ pkg/

ARG version
# build
RUN CGO_ENABLED=1 go build -ldflags "-X 'github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/version.Version=$version' -linkmode external -w -extldflags \"-static\"" -tags netgo -a -o clusters-service main.go

FROM docker.io/governmentpaas/git-ssh

ENV CS_PATH=cmd/clusters-service

# # Add git provider hostnames to known hosts file so we can use
# # StrickHostKeyChecking with git+ssh
COPY ./tools/known_hosts.sh /tmp/known_hosts.sh
RUN bash /tmp/known_hosts.sh /etc/ssh/ssh_known_hosts && rm /tmp/known_hosts.sh

RUN apk add --no-cache ca-certificates tini

COPY --from=builder /workspace/$CS_PATH/clusters-service /usr/local/bin/

RUN addgroup -S clusters-service && adduser -S clusters-service -G clusters-service

#
# FIXME: it would be nice not run as root but k8s mounts the sqlite host volume as root:root
# and then the clusters-service user can't write to it.
#
# USER clusters-service
# RUN mkdir -p /home/clusters-service/.ssh

ENTRYPOINT [ "/sbin/tini", "--", "clusters-service" ]
