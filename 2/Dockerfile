FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /url-shortener ./cmd/server

FROM alpine:latest
WORKDIR /

COPY --from=builder /url-shortener /url-shortener

COPY ./ui /ui

EXPOSE 8080

CMD ["/url-shortener"]
