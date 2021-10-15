from golang:1.17.1 as builder

ENV CS_PATH=cmd/clusters-service

WORKDIR /workspace/$CS_PATH

# copy in other module dependencies
COPY common/ /workspace/common/

# copy modules manifests
COPY $CS_PATH/go.mod go.mod
COPY $CS_PATH/go.sum go.sum

ENV GITHUB_BUILD_USERNAME=wge-build-bot
ARG GITHUB_BUILD_TOKEN
RUN git config --global url."https://${GITHUB_BUILD_USERNAME}:${GITHUB_BUILD_TOKEN}@github.com".insteadOf "https://github.com"

# cache modules
RUN go mod download

# copy source code
COPY $CS_PATH/main.go main.go
COPY $CS_PATH/api/ api/
COPY $CS_PATH/app/ app/
COPY $CS_PATH/pkg/ pkg/

ARG version
# build
RUN CGO_ENABLED=1 go build -ldflags "-X 'github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/version.Version=$version' -linkmode external -w -extldflags \"-static\"" -a -o clusters-service main.go

FROM docker.io/governmentpaas/git-ssh

ENV CS_PATH=cmd/clusters-service

# # Add git provider hostnames to known hosts file so we can use
# # StrickHostKeyChecking with git+ssh
# ADD ./known_hosts.sh /home/wkp/known_hosts.sh
# RUN bash /home/wkp/known_hosts.sh /etc/ssh/ssh_known_hosts && rm /home/wkp/known_hosts.sh

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
