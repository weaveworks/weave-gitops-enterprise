FROM node:14-bullseye AS build
USER node
WORKDIR /home/node
COPY --chown=node:node package.json /home/node
COPY --chown=node:node .npmrc /home/node
COPY --chown=node:node yarn.lock /home/node
COPY --chown=node:node Makefile /home/node

ARG GITHUB_TOKEN
RUN make node_modules

COPY --chown=node:node . /home/node
ARG version
RUN make build VERSION=${version}

FROM alpine:3.9
RUN apk add --no-cache bash openssh-client

WORKDIR /
COPY --from=build /home/node/build /html

ARG now
ARG revision
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
