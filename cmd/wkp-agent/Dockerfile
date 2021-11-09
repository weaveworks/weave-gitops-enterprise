FROM docker.io/governmentpaas/git-ssh

ARG now
ARG revision
ARG version
LABEL maintainer="Weaveworks <help@weave.works>" \
      org.opencontainers.image.created="${now}" \
      org.opencontainers.image.description="The Weave GitOps Enterprise WKP Agent" \
      org.opencontainers.image.documentation="https://docs.gitops.weave.works/" \
      org.opencontainers.image.licenses="MPL-2.0" \
      org.opencontainers.image.revision="${revision}" \
      org.opencontainers.image.source="https://github.com/weaveworks/weave-gitops-enterprise" \
      org.opencontainers.image.title="wkp-agent" \
      org.opencontainers.image.url="https://github.com/weaveworks/weave-gitops-enterprise" \
      org.opencontainers.image.vendor="Weaveworks" \
      org.opencontainers.image.version="${version}"

COPY wkp-agent /

ENTRYPOINT ["/wkp-agent"]

CMD ["--watch"]