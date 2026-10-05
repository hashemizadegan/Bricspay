import { cookies } from "next/headers";

export type VerificationStatus =
  | "pending_submission"
  | "pending_review"
  | "verified"
  | "rejected";

export type VerificationState = {
  authenticated: true;
  status: VerificationStatus;
};

const validStatuses: readonly VerificationStatus[] = [
  "pending_submission",
  "pending_review",
  "verified",
  "rejected",
];

export async function getVerificationState(): Promise<VerificationState | null> {
  const endpoint = process.env.KYC_STATUS_URL;
  if (!endpoint) throw new Error("KYC_STATUS_URL is not configured");

  const cookieStore = await cookies();
  const cookieHeader = cookieStore.toString();
  const headers = new Headers();
  if (cookieHeader) headers.set("Cookie", cookieHeader);
  const response = await fetch(endpoint, {
    method: "GET",
    headers,
    cache: "no-store",
  });

  if (response.status === 401) return null;
  if (!response.ok) {
    throw new Error(`KYC status endpoint returned ${response.status}`);
  }

  const payload: unknown = await response.json();
  if (!payload || typeof payload !== "object") {
    throw new Error("Invalid KYC status response");
  }
  const data = payload as Record<string, unknown>;
  if (data.authenticated === false) return null;
  if (data.authenticated !== true ||
      typeof data.status !== "string" ||
      !validStatuses.includes(data.status as VerificationStatus)) {
    throw new Error("Invalid KYC status response");
  }

  return {
    authenticated: true,
    status: data.status as VerificationStatus,
  };
}
