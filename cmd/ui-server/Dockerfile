FROM alpine:3.9
RUN apk add --no-cache bash openssh-client

WORKDIR /
COPY html /html/
COPY ui-server /

ARG revision
LABEL maintainer="Weaveworks <help@weave.works>" \
      org.opencontainers.image.title="ui-server" \
      org.opencontainers.image.source="https://github.com/weaveworks/wks/tree/master/ui" \
      org.opencontainers.image.revision="${revision}" \
      org.opencontainers.image.vendor="Weaveworks"

ENTRYPOINT ["/ui-server"]
EXPOSE 80
