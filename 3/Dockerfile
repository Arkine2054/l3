# Stage 1: сборка Go
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Собираем бинарь
RUN go build -o server ./cmd/server

# Stage 2: финальный образ
FROM alpine:latest

WORKDIR /app

# Копируем бинарь
COPY --from=builder /app/server .

# Копируем веб-файлы
COPY web/templates ./web/templates
COPY web/static ./web/static

# Копируем миграции
COPY internal/migrations ./migrations

# Открываем порт
EXPOSE 8080

# Запуск приложения
ENTRYPOINT ["./server"]
