package db

import "database/sql"

func MigrateKYC(db *sql.DB) error {
	_, err := db.Exec(`
DO $$ BEGIN
  CREATE TYPE entity_type AS ENUM ('INDIVIDUAL', 'CORPORATE', 'BANK');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
DO $$ BEGIN
  CREATE TYPE user_role AS ENUM ('user', 'admin');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
DO $$ BEGIN
  CREATE TYPE kyc_status AS ENUM ('PENDING', 'SUBMITTED', 'APPROVED', 'REJECTED');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
DO $$ BEGIN
  CREATE TYPE doc_type AS ENUM (
    'PASSPORT_SIGNATORY', 'COMMERCIAL_REGISTRY', 'STATUTES',
    'BANK_LICENSE', 'TAX_CERTIFICATE'
  );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS kyc_profiles (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL UNIQUE REFERENCES users(id),
  entity_type entity_type NOT NULL DEFAULT 'INDIVIDUAL',
  status kyc_status NOT NULL DEFAULT 'PENDING',
  rejection_reason TEXT,
  resubmit_count INT NOT NULL DEFAULT 0,
  submitted_at TIMESTAMPTZ,
  reviewed_at TIMESTAMPTZ,
  reviewer_id BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS kyc_documents (
  id BIGSERIAL PRIMARY KEY,
  profile_id BIGINT NOT NULL REFERENCES kyc_profiles(id),
  doc_type doc_type NOT NULL,
  filename TEXT NOT NULL,
  content_type TEXT,
  size_bytes BIGINT,
  sha256 TEXT,
  storage_path TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE kyc_profiles ADD COLUMN IF NOT EXISTS rejection_reason TEXT;
ALTER TABLE kyc_profiles ADD COLUMN IF NOT EXISTS resubmit_count INT NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone TEXT;
`)
	return err
}
