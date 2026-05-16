FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/shortener ./cmd/shortener

FROM alpine:3.20

WORKDIR /app

RUN adduser -D -H shortener

COPY --from=builder /bin/shortener /app/shortener
COPY migrations /app/migrations

ENV HTTP_ADDR=:8080
ENV MIGRATIONS_DIR=/app/migrations

USER shortener
EXPOSE 8080

CMD ["/app/shortener"]
