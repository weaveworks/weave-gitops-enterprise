FROM alpine

ENV GITOPS_JWT_ENCRYPTION_SECRET=supersecret

COPY ./cmd/clusters-service/bin /app
