FROM golang:1.27.3-bookworm AS build
ENV CGO_ENABLED=0 GOTOOLCHAIN=local
RUN go install -tags='no_clickhouse no_libsql no_mssql no_mysql no_sqlite3 no_vertica no_ydb no_azuresql' github.com/pressly/goose/v3/cmd/goose@v3.28.0

FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab
ARG APP_VERSION=0.0.1
LABEL org.opencontainers.image.version=${APP_VERSION}
COPY --from=build /go/bin/goose /usr/local/bin/goose
COPY apps/api/db/migrations /migrations
USER 65532:65532
ENTRYPOINT ["goose", "-dir", "/migrations"]
CMD ["up"]
