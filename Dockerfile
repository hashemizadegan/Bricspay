# -------------------------------------------------------------
# Stage 1: Build the Go binary
# -------------------------------------------------------------
FROM golang:1.23-alpine AS builder

WORKDIR /app

# نصب ابزارهای مورد نیاز برای وابستگی‌های پایه
RUN apk add --no-cache git ca-certificates

# کپی فایل‌های مدیریت وابستگی
COPY go.mod go.sum* ./

# دانلود وابستگی‌ها و اطمینان از صحت ماژول‌ها
RUN go mod download

# کپی کل کدهای پروژه
COPY . .

# کامپایل مستقل و استاتیک باینری
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-w -s" -o bricspay-api ./cmd/api

# -------------------------------------------------------------
# Stage 2: Final minimal runtime image
# -------------------------------------------------------------
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

# ساخت دایرکتوری ذخیره‌سازی مدارک KYC
RUN mkdir -p /data/uploads && chmod 700 /data/uploads

# کپی باینری کامپایل‌شده از مرحله بیلد
COPY --from=builder /app/bricspay-api .

# کپی فایل‌های فرانت‌اند و استاتیک (در صورت وجود)
COPY --from=builder /app/static ./static

EXPOSE 8080

CMD ["./bricspay-api"]
