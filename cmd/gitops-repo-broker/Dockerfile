FROM docker.io/governmentpaas/git-ssh

ARG now
ARG revision
ARG version
LABEL maintainer="Weaveworks <help@weave.works>" \
      org.opencontainers.image.created="${now}" \
      org.opencontainers.image.description="The Weave GitOps Enterprise GitOps Repo Broker" \
      org.opencontainers.image.documentation="https://docs.gitops.weave.works/" \
      org.opencontainers.image.licenses="MPL-2.0" \
      org.opencontainers.image.revision="${revision}" \
      org.opencontainers.image.source="https://github.com/weaveworks/weave-gitops-enterprise" \
      org.opencontainers.image.title="weave-gitops-enterprise-gitops-repo-broker" \
      org.opencontainers.image.url="https://github.com/weaveworks/weave-gitops-enterprise" \
      org.opencontainers.image.vendor="Weaveworks" \
      org.opencontainers.image.version="${version}"

# Add git provider hostnames to known hosts file so we can use
# StrickHostKeyChecking with git+ssh
ADD ./known_hosts.sh /home/wkp/known_hosts.sh
RUN bash /home/wkp/known_hosts.sh /etc/ssh/ssh_known_hosts && \
    rm /home/wkp/known_hosts.sh

RUN wget https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.11.0/kubeseal-linux-amd64 -O kubeseal && \
    echo 'ae580e28400b4aa806adf3fd6b002799763820927bd6ac5a5df5250d50377e67  kubeseal' | sha256sum -c && \
    install -m 755 kubeseal /usr/local/bin/kubeseal

ADD gitops-repo-broker /
ENTRYPOINT ["/gitops-repo-broker"]
