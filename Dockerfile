FROM golang:1.24.2-alpine3.21 AS builder
WORKDIR /app
COPY . .

RUN go build cmd/server/main.go

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /app/main .
COPY ./config/app.env ./config/app.env
EXPOSE 8080
CMD ["/app/main"]