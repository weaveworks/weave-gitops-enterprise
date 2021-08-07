from golang:1.16.6 as builder

ENV CS_PATH=cmd/capi-server

WORKDIR /workspace/$CS_PATH

# copy in other module dependencies
COPY common/ /workspace/common/

# copy modules manifests
COPY $CS_PATH/go.mod go.mod
COPY $CS_PATH/go.sum go.sum

# cache modules
RUN go mod download

# copy source code
COPY $CS_PATH/main.go main.go
COPY $CS_PATH/api/ api/
COPY $CS_PATH/app/ app/
COPY $CS_PATH/pkg/ pkg/

ARG version
# build
RUN CGO_ENABLED=1 go build -ldflags "-X 'github.com/weaveworks/wks/cmd/capi-server/pkg/version.Version=$version' -linkmode external -w -extldflags \"-static\"" -a -o capi-server main.go

FROM docker.io/governmentpaas/git-ssh

ENV CS_PATH=cmd/capi-server

# # Add git provider hostnames to known hosts file so we can use
# # StrickHostKeyChecking with git+ssh
# ADD ./known_hosts.sh /home/wkp/known_hosts.sh
# RUN bash /home/wkp/known_hosts.sh /etc/ssh/ssh_known_hosts && rm /home/wkp/known_hosts.sh

RUN apk add --no-cache ca-certificates tini

COPY --from=builder /workspace/$CS_PATH/capi-server /usr/local/bin/

RUN addgroup -S capi-server && adduser -S capi-server -G capi-server

#
# FIXME: it would be nice not run as root but k8s mounts the sqlite host volume as root:root
# and then the capi-server user can't write to it.
#
# USER capi-server
# RUN mkdir -p /home/capi-server/.ssh

ENTRYPOINT [ "/sbin/tini", "--", "capi-server" ]
