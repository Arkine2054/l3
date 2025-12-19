FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o api ./cmd/api


FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/api ./api
COPY --from=builder /app/web ./web
COPY --from=builder /app/internal/migrations ./internal/migrations

EXPOSE 8080

CMD ["./api"]
