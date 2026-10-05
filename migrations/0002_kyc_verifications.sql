-- 0002_kyc_verifications.sql
-- جدول احراز هویت (Identity Verification) که kyc_handlers.go انتظارش را دارد.
-- مستقل از kyc_profiles / kyc_documents می‌ماند تا سیستم اسناد آپلودی دست‌نخورده بماند.

CREATE TABLE IF NOT EXISTS kyc_verifications (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL UNIQUE REFERENCES users(id),
    full_name       TEXT NOT NULL,
    document_type   TEXT NOT NULL,
    document_number TEXT NOT NULL,
    country         TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'approved', 'rejected')),
    notes           TEXT,
    submitted_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_kyc_verifications_status
    ON kyc_verifications (status);
