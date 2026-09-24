FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bricspay-api ./cmd/api

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /bricspay-api .

EXPOSE 8080
CMD ["./bricspay-api"]
