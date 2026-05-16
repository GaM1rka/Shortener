FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/shortener ./cmd/shortener

FROM alpine:3.20

WORKDIR /app

RUN adduser -D -H shortener

COPY --from=builder /bin/shortener /app/shortener
COPY backend/migrations /app/migrations
COPY frontend /app/frontend

ENV HTTP_ADDR=:8080
ENV MIGRATIONS_DIR=/app/migrations
ENV FRONTEND_DIR=/app/frontend

USER shortener
EXPOSE 8080

CMD ["/app/shortener"]
