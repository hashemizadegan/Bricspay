CREATE TABLE IF NOT EXISTS accounts (
    id       BIGSERIAL PRIMARY KEY,
    code     TEXT NOT NULL UNIQUE,
    currency TEXT NOT NULL,
    balance  BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0)
);

CREATE TABLE IF NOT EXISTS wallets (
    id       BIGSERIAL PRIMARY KEY,
    currency TEXT NOT NULL,
    balance  BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0)
);

CREATE TABLE IF NOT EXISTS transactions (
    id              BIGSERIAL PRIMARY KEY,
    from_wallet     BIGINT NOT NULL REFERENCES wallets(id),
    to_wallet       BIGINT NOT NULL REFERENCES wallets(id),
    amount          BIGINT NOT NULL CHECK (amount > 0),
    currency        TEXT NOT NULL,
    status          TEXT NOT NULL,
    reference       TEXT,
    idempotency_key TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_tx_idem UNIQUE (idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_tx_from ON transactions(from_wallet);
CREATE INDEX IF NOT EXISTS idx_tx_to   ON transactions(to_wallet);
