const schema = `
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL DEFAULT 'user',
    entity_type   TEXT NOT NULL DEFAULT 'CORPORATE',
    status        TEXT NOT NULL DEFAULT 'PENDING',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS wallets (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    currency   TEXT NOT NULL DEFAULT 'USD',
    balance    NUMERIC(20,8) NOT NULL DEFAULT 0 CHECK (balance >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, currency)
);

CREATE TABLE IF NOT EXISTS transactions (
    id          BIGSERIAL PRIMARY KEY,
    from_wallet BIGINT REFERENCES wallets(id),
    to_wallet   BIGINT REFERENCES wallets(id),
    amount      NUMERIC(20,8) NOT NULL CHECK (amount > 0),
    currency    TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'pending',
    reference   TEXT NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transactions_from
    ON transactions(from_wallet);

CREATE INDEX IF NOT EXISTS idx_transactions_to
    ON transactions(to_wallet);

CREATE TABLE IF NOT EXISTS kyc_profiles (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             BIGINT NOT NULL UNIQUE
                        REFERENCES users(id) ON DELETE CASCADE,
    legal_name          TEXT NOT NULL,
    registration_number TEXT NOT NULL,
    tax_id              TEXT NOT NULL,
    jurisdiction        TEXT NOT NULL,
    contact_phone       TEXT NOT NULL,
    risk_tier           TEXT NOT NULL DEFAULT 'MEDIUM',
    status              TEXT NOT NULL DEFAULT 'SUBMITTED',
    reviewer_notes      TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS kyc_documents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id      UUID NOT NULL
                    REFERENCES kyc_profiles(id) ON DELETE CASCADE,
    document_type   TEXT NOT NULL,
    file_path       TEXT NOT NULL,
    original_name   TEXT NOT NULL,
    checksum_sha256 TEXT NOT NULL,
    size_bytes      BIGINT NOT NULL CHECK (size_bytes >= 0),
    status          TEXT NOT NULL DEFAULT 'PENDING',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id            BIGSERIAL PRIMARY KEY,
    actor_id      BIGINT,
    actor_email   TEXT NOT NULL,
    action        TEXT NOT NULL,
    target_entity TEXT NOT NULL,
    target_id     TEXT,
    ip_address    TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_kyc_profiles_user_id
    ON kyc_profiles(user_id);

CREATE INDEX IF NOT EXISTS idx_kyc_documents_profile
    ON kyc_documents(profile_id);

CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at
    ON audit_logs(created_at);
`
