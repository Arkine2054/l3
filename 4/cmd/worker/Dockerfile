FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o worker ./cmd/worker


FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/worker ./worker
COPY --from=builder /app/internal/assets ./internal/assets
COPY --from=builder /app/internal/migrations ./internal/migrations

RUN mkdir -p /app/data/original /app/data/processed /app/data/thumbs

CMD ["./worker"]
