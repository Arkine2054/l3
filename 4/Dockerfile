# =========================
# Stage 1: Build Go binary
# =========================
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Собираем бинарник
RUN go build -o server ./cmd/server

# =========================
# Stage 2: Runtime image
# =========================
FROM alpine:latest

# Нужны для работы HTTPS и некоторых функций Go
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Копируем собранный бинарник
COPY --from=builder /app/server .

# Копируем веб-ресурсы
COPY web ./web

# Копируем шрифт и миграции
COPY internal/assets ./internal/assets
COPY internal/migrations ./internal/migrations

# Создаём директории для хранения изображений
RUN mkdir -p /app/data/original /app/data/processed /app/data/thumbs

# Порты
ENV PORT=8080
EXPOSE 8080

# Точка входа
CMD ["./server"]
