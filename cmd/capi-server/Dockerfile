from golang:1.16.4 as builder

WORKDIR /capi-server

# copy modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# cache modules
RUN go mod download

# copy source code
COPY main.go main.go
COPY api/ api/
COPY app/ app/
COPY pkg/ pkg/

# build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o capi-server main.go

FROM docker.io/governmentpaas/git-ssh

# # Add git provider hostnames to known hosts file so we can use
# # StrickHostKeyChecking with git+ssh
# ADD ./known_hosts.sh /home/wkp/known_hosts.sh
# RUN bash /home/wkp/known_hosts.sh /etc/ssh/ssh_known_hosts && rm /home/wkp/known_hosts.sh

RUN apk add --no-cache ca-certificates tini

COPY --from=builder /capi-server/capi-server /usr/local/bin/

RUN addgroup -S capi-server && adduser -S capi-server -G capi-server

USER capi-server

RUN mkdir -p /home/capi-server/.ssh

ENTRYPOINT [ "/sbin/tini", "--", "capi-server" ]
