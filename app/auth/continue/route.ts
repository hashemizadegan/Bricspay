import { NextRequest, NextResponse } from "next/server";
import { getVerificationState } from "@/lib/verification";

function safeReturnTo(value: string | null): string {
  if (!value || !value.startsWith("/") || value.startsWith("//") || value.includes("\\")) {
    return "/dashboard";
  }
  return value;
}

export async function GET(request: NextRequest) {
  const returnTo = safeReturnTo(request.nextUrl.searchParams.get("returnTo"));
  const state = await getVerificationState();
  if (!state) {
    const login = new URL("/login", request.url);
    login.searchParams.set("returnTo", returnTo);
    return NextResponse.redirect(login);
  }

  if (state.status === "verified") {
    return NextResponse.redirect(new URL(returnTo, request.url));
  }

  const verification = new URL("/settings/id_verification", request.url);
  verification.searchParams.set("returnTo", returnTo);
  return NextResponse.redirect(verification);
}
