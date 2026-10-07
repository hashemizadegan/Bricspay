-- 0003_kyc_document_blobs.sql
-- Stores uploaded KYC document binaries in Postgres (BYTEA)

CREATE TABLE IF NOT EXISTS kyc_document_blobs (
    id TEXT PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    document_type TEXT NOT NULL,
    original_name TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size_bytes BIGINT NOT NULL CHECK (size_bytes >= 0),
    sha256 TEXT NOT NULL,
    content BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_kyc_document_blobs_user_id
    ON kyc_document_blobs(user_id);

CREATE INDEX IF NOT EXISTS idx_kyc_document_blobs_sha256
    ON kyc_document_blobs(sha256);
