# مرحله اول: بیلد پروژه با Go 1.23
FROM golang:1.23-alpine AS builder

WORKDIR /app

# کپی فایلهای وابستگی و دانلود آنها
COPY go.mod go.sum ./
RUN go mod download

# کپی کل سورس کد
COPY . .

# کامپایل باینری Go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bricspay-api ./cmd/api

# مرحله دوم: ایمیج نهایی و سبک زمان اجرا
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# کپی باینری کامپایل‌شده
COPY --from=builder /app/bricspay-api .

# کپی فایل‌های استاتیک فرانت‌اند بر اساس ساختار پروژه
COPY --from=builder /app/internal/api/static ./internal/api/static

EXPOSE 8080

CMD ["./bricspay-api"]
