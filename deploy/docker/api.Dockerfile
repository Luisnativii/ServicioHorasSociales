FROM golang:1.27.3-bookworm AS dev
WORKDIR /workspace/apps/api
ENV GOTOOLCHAIN=local GOFLAGS=-mod=readonly
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api/ ./
COPY scripts/dev-api.sh /workspace/scripts/dev-api.sh
CMD ["bash", "/workspace/scripts/dev-api.sh"]

FROM dev AS build
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab AS runtime
ARG APP_VERSION=0.0.1
LABEL org.opencontainers.image.version=${APP_VERSION}
COPY --from=build /out/api /api
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/api"]
