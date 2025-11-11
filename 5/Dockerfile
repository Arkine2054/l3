# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Устанавливаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Собираем бинарь
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Финальный образ
FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/server /app/server
COPY --from=builder /app/web /app/web
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8080

ENV DATABASE_URL=postgres://postgres:postgres@db:5432/events?sslmode=disable
ENV BOOKING_EXPIRATION_MINUTES=2

CMD ["/app/server"]
