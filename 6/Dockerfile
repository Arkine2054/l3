# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Зависимости
COPY go.mod go.sum ./
RUN go mod download

# Исходники
COPY . .

# Сборка бинарника
RUN go build -o server ./cmd/server

# --- Финальный минимальный образ ---
FROM alpine:latest

WORKDIR /app

# Устанавливаем зависимости (если нужно для PostgreSQL)
RUN apk add --no-cache ca-certificates tzdata

# Копируем бинарник из builder
COPY --from=builder /app/server .

# Опционально: если нужно копировать html или шаблоны
COPY web ./web

# Запускаем приложение
CMD ["./server"]
