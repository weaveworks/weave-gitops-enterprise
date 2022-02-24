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

COPY cmd/clusters-service/clusters-service /usr/local/bin

RUN addgroup -S clusters-service && adduser -S clusters-service -G clusters-service

#
# FIXME: it would be nice not run as root but k8s mounts the sqlite host volume as root:root
# and then the clusters-service user can't write to it.
#
# USER clusters-service
# RUN mkdir -p /home/clusters-service/.ssh

ENTRYPOINT [ "/sbin/tini", "--", "clusters-service" ]
