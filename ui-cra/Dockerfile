FROM alpine:3.9
RUN apk add --no-cache bash openssh-client

WORKDIR /
COPY build /html

ARG now
ARG revision
ARG version
LABEL maintainer="Weaveworks <help@weave.works>" \
      org.opencontainers.image.title="ui-server" \
      org.opencontainers.image.source="https://github.com/weaveworks/weave-gitops-enterprise" \
      org.opencontainers.image.revision="${revision}" \
      org.opencontainers.image.vendor="Weaveworks" \
      org.opencontainers.image.created="${now}" \
      org.opencontainers.image.description="The Weave GitOps Enterprise UI Server" \
      org.opencontainers.image.documentation="https://docs.gitops.weave.works/" \
      org.opencontainers.image.url="https://docs.gitops.weave.works/" \
      org.opencontainers.image.version="${version}"

CMD ["sh", "-c", "mv -v /html/* /target/"]
