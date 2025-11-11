# -------- Build Stage --------
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Собираем бинарник
RUN go build -o /url-shortener ./cmd/server

# -------- Final Stage --------
FROM alpine:latest
WORKDIR /

# Копируем бинарник
COPY --from=builder /url-shortener /url-shortener

# Копируем UI
COPY ./ui /ui

# Порты
EXPOSE 8080

# Запуск приложения
CMD ["/url-shortener"]
