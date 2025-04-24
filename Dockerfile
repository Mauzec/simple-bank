FROM golang:1.24.2-alpine3.21 AS builder
WORKDIR /app
COPY . .
RUN go build cmd/server/main.go
RUN apk add curl
RUN mkdir -p /tmp/migrate_download \
    && curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.2/migrate.linux-amd64.tar.gz \
    | tar -xz -C /tmp/migrate_download \
    && mv /tmp/migrate_download/migrate ./migrate_bin \
    && rm -rf /tmp/migrate_download



FROM alpine:3.21
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/migrate_bin /usr/bin/migrate
RUN rm -rf /app/migrate_bin
COPY db/migrate ./migrate
COPY ./config/app.env ./config/app.env
COPY start.sh .
COPY wait-for.sh .
EXPOSE 8080
ENTRYPOINT ["/app/start.sh"]
CMD ["/app/main"]
