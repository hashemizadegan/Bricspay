# بسته هدایت احراز هویت و KYC برای Next.js App Router

این بسته چند فایل مستقل TypeScript/TSX است؛ drop-in یا یکپارچه‌شده با سیستم ورود شما نیست. پس از افزودن فایل‌ها، آدرس و ساختار API و handlerهای موجود ورود/ثبت‌نام را مطابق مراحل زیر متصل کنید.

## فایل‌ها

- `lib/verification.ts`: خواندن کوکی درخواست جاری در سرور و دریافت وضعیت KYC.
- `app/auth/continue/route.ts`: هدایت کاربر پس از ورود یا ثبت‌نام.
- `app/settings/id_verification/page.tsx`: صفحه فارسی راست‌چین وضعیت احراز هویت.

## مراحل اتصال

1. این فایل‌ها را در همان مسیرهای پروژه Next.js دارای App Router قرار دهید. پروژه باید از TypeScript و `next` استفاده کند؛ این بسته `package.json` یا وابستگی اضافه ندارد.
2. متغیر `KYC_STATUS_URL` را در محیط سرور تنظیم کنید. مقدار آن باید URL کامل endpoint وضعیت باشد. این متغیر در کد کلاینت استفاده نمی‌شود.
3. endpoint باید کوکی نشست را از هدر `Cookie` دریافت کند و JSON مطابق قرارداد پایین برگرداند. endpoint باید برای نشست نامعتبر پاسخ `401` بدهد یا JSON با `authenticated:false` برگرداند.
4. در handler موفقیت موجود ورود و ثبت‌نام، کاربر را به `/auth/continue?returnTo=...` هدایت کنید. مقدار `returnTo` را URL-encode کنید؛ برای نمونه `/auth/continue?returnTo=%2Forders`.
5. مسیرهای `/login`، `/dashboard`، `/kyc/start` و `/kyb/start` باید در برنامه وجود داشته باشند یا لینک‌ها را متناسب با مسیرهای واقعی خودتان تغییر دهید.
6. متناسب با نسخه Next.js نصب‌شده، امضای APIهای `cookies()` و `NextResponse` را بررسی کنید. این پیاده‌سازی از API ناهمگام `cookies()` در نسخه‌های جدید Next.js استفاده می‌کند.

## قرارداد API وضعیت KYC

در درخواست سرور به `KYC_STATUS_URL`، کوکی ورودی مرورگر بدون بازسازی نشست در هدر `Cookie` فرستاده می‌شود. API باید برای نشست احراز‌شده پاسخ JSON زیر را بدهد:

```json
{"authenticated":true,"status":"pending_submission"}
```

مقادیر مجاز `status`:

- `pending_submission`
- `pending_review`
- `verified`
- `rejected`

برای نشست نامعتبر پاسخ HTTP `401` یا `{ "authenticated": false }` برگردانید. هر پاسخ موفق دیگری که schema یا status معتبر نداشته باشد، خطای API محسوب می‌شود و در این پیاده‌سازی exception ایجاد می‌کند؛ آن را با error handling پروژه مدیریت کنید. API باید از same-origin یا endpoint قابل‌دسترسی از سرور برنامه باشد.

## رفتار هدایت و امنیت

- کاربر بدون نشست به `/login` هدایت می‌شود.
- کاربر تأییدشده به `returnTo` داخلی معتبر یا `/dashboard` می‌رود.
- سایر وضعیت‌ها به `/settings/id_verification` هدایت می‌شوند و `returnTo` امن در query حفظ می‌شود.
- `returnTo` فقط وقتی پذیرفته می‌شود که با یک `/` آغاز شود؛ `//`، بک‌اسلش و مسیر فاقد `/` آغازین رد می‌شوند. مقصد بیرونی پذیرفته نمی‌شود.
- لینک‌های صفحه به `/kyc/start` و `/kyb/start` می‌روند. محافظت از مسیرها و authorization هر جریان باید در خود برنامه انجام شود.
