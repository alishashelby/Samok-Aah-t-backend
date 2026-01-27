FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

FROM alpine
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY migrate.sh /app/migrate.sh
COPY sql /app/sql
RUN chmod +x /app/migrate.sh
WORKDIR /app

CMD ["/app/migrate.sh"]