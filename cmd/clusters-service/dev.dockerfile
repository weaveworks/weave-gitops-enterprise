FROM alpine:3.18

RUN apk --no-cache add ca-certificates gnupg \
  && update-ca-certificates

ENV GITOPS_JWT_ENCRYPTION_SECRET=supersecret

COPY ./cmd/clusters-service/bin /app
