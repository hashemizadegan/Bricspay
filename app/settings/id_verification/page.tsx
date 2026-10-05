import { redirect } from "next/navigation";
import { getVerificationState } from "@/lib/verification";

type PageProps = {
  searchParams: Promise<{ returnTo?: string }>;
};

function safeReturnTo(value: string | undefined): string {
  if (!value || !value.startsWith("/") || value.startsWith("//") || value.includes("\\")) {
    return "/dashboard";
  }
  return value;
}

const copy = {
  pending_submission: {
    title: "تکمیل احراز هویت",
    message: "برای دسترسی به ادامه خدمات، اطلاعات هویتی خود را ثبت کنید.",
    action: "شروع احراز هویت شخصی",
  },
  pending_review: {
    title: "مدارک شما در حال بررسی است",
    message: "درخواست شما دریافت شده است. پس از بررسی نتیجه در همین بخش نمایش داده می‌شود.",
    action: "مشاهده وضعیت درخواست",
  },
  rejected: {
    title: "نیاز به اصلاح مدارک",
    message: "درخواست احراز هویت تأیید نشده است. اطلاعات را بررسی و دوباره ارسال کنید.",
    action: "اصلاح و ارسال دوباره",
  },
} as const;

export default async function IdVerificationPage({ searchParams }: PageProps) {
  const params = await searchParams;
  const returnTo = safeReturnTo(params.returnTo);
  const state = await getVerificationState();

  if (!state) {
    const login = new URLSearchParams({ returnTo: "/settings/id_verification" });
    redirect(`/login?${login.toString()}`);
  }
  if (state.status === "verified") redirect(returnTo);

  const content = copy[state.status];
  const kycHref = "/kyc/start";

  return (
    <main lang="fa" dir="rtl" style={{ maxWidth: 680, margin: "64px auto", padding: "0 24px", fontFamily: "sans-serif", lineHeight: 1.9 }}>
      <h1>{content.title}</h1>
      <p>{content.message}</p>
      <nav aria-label="روش احراز هویت" style={{ display: "flex", flexWrap: "wrap", gap: 16, marginTop: 24 }}>
        <a href={kycHref}>{content.action}</a>
        <a href="/kyb/start">احراز هویت کسب‌وکار</a>
      </nav>
      <p style={{ marginTop: 32 }}>
        <a href={returnTo}>بازگشت به ادامه کار</a>
      </p>
    </main>
  );
}
