FROM golang:1.23-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /usr/local/bin/worker ./cmd/worker

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /usr/local/bin/worker /usr/local/bin/worker
ENTRYPOINT ["/usr/local/bin/worker"]
