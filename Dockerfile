# مرحله اول: بیلد با Go
FROM golang:1.22-alpine AS builder

WORKDIR /app

# مدیریت کش پیش‌نیازها
COPY go.mod go.sum ./
RUN go mod download

# کپی سورس‌کد
COPY . .

# کامپایل باینری به صورت استاتیک
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bricspay-api ./cmd/api

# مرحله دوم: ایمیج نهایی بسیار سبک
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# کپی باینری و پوشه استاتیک از مرحله قبل
COPY --from=builder /app/bricspay-api .
COPY --from=builder /app/static ./static

EXPOSE 8080

CMD ["./bricspay-api"]
