FROM alpine:3.12

ADD event-writer /

RUN apk add --no-cache tini

ENTRYPOINT [ "/sbin/tini", "--", "/event-writer" ]
