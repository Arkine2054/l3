FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /usr/local/bin/api ./cmd/api

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /usr/local/bin/api /usr/local/bin/api

COPY ui ./ui
COPY .env .env

EXPOSE 8083
ENTRYPOINT ["/usr/local/bin/api"]
