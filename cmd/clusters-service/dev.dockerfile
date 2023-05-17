FROM gcr.io/distroless/static:nonroot

ENV GITOPS_JWT_ENCRYPTION_SECRET=supersecret

COPY ./cmd/clusters-service/bin /app
