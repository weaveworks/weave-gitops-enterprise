# syntax = docker/dockerfile:1.2
FROM node:16-bullseye AS build
USER node
WORKDIR /home/node
COPY --chown=node:node package.json /home/node
COPY --chown=node:node .npmrc /home/node
COPY --chown=node:node yarn.lock /home/node
COPY --chown=node:node Makefile /home/node
COPY --chown=node:node tsconfig.json /home/node

ARG REACT_APP_DISABLE_PROGRESSIVE_DELIVERY
ARG GITHUB_TOKEN
RUN --mount=type=cache,target=/home/node/node_modules,uid=1000,gid=1000 make node_modules

COPY --chown=node:node . /home/node/
ARG version
# When set to true, Create React App treats warnings as failures in the build.
# It also makes the test runner non-watching. Most CIs set this flag by default.
# https://create-react-app.dev/docs/advanced-configuration
ENV CI=true
RUN --mount=type=cache,target=/home/node/node_modules,uid=1000,gid=1000 make build VERSION=${version}

FROM alpine:3.16

RUN apk upgrade libssl1.1 libcrypto1.1

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
