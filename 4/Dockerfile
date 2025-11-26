FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o server ./cmd/server

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/server .

COPY web ./web

COPY internal/assets ./internal/assets
COPY internal/migrations ./internal/migrations

RUN mkdir -p /app/data/original /app/data/processed /app/data/thumbs

ENV PORT=8080
EXPOSE 8080

CMD ["./server"]
