package db

import "database/sql"

// MigrateKYC creates all KYC-related tables if they do not exist.
func MigrateKYC(db *sql.DB) error {
	_, err := db.Exec(`
		DO $$ BEGIN
			CREATE TYPE entity_type_enum AS ENUM ('CORPORATE', 'BANK');
		EXCEPTION WHEN duplicate_object THEN NULL;
		END $$;

		DO $$ BEGIN
			CREATE TYPE user_role_enum AS ENUM ('member', 'admin');
		EXCEPTION WHEN duplicate_object THEN NULL;
		END $$;

		DO $$ BEGIN
			CREATE TYPE kyc_status_enum AS ENUM ('SUBMITTED', 'UNDER_REVIEW', 'APPROVED', 'REJECTED');
		EXCEPTION WHEN duplicate_object THEN NULL;
		END $$;

		DO $$ BEGIN
			CREATE TYPE doc_type_enum AS ENUM (
				'COMMERCIAL_REGISTRY', 'STATUTES', 'PASSPORT_SIGNATORY',
				'BANK_LICENSE', 'TAX_CERTIFICATE', 'OFAC_SCREENING'
			);
		EXCEPTION WHEN duplicate_object THEN NULL;
		END $$;

		CREATE TABLE IF NOT EXISTS users (
			id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email           TEXT            NOT NULL UNIQUE,
			password_hash   TEXT            NOT NULL,
			entity_type     entity_type_enum NOT NULL,
			role            user_role_enum   NOT NULL DEFAULT 'member',
			status          kyc_status_enum  NOT NULL DEFAULT 'SUBMITTED',
			created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
			updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS kyc_profiles (
			id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id             UUID            NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
			legal_name          TEXT            NOT NULL,
			registration_number TEXT            NOT NULL,
			tax_id              TEXT            NOT NULL,
			jurisdiction        TEXT            NOT NULL,
			contact_phone       TEXT            NOT NULL,
			risk_tier           TEXT            NOT NULL DEFAULT 'MEDIUM',
			status              kyc_status_enum  NOT NULL DEFAULT 'SUBMITTED',
			reviewer_notes      TEXT,
			created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
			updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS kyc_documents (
			id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			profile_id      UUID            NOT NULL REFERENCES kyc_profiles(id) ON DELETE CASCADE,
			document_type   doc_type_enum    NOT NULL,
			file_path       TEXT            NOT NULL,
			original_name   TEXT            NOT NULL,
			checksum_sha256 TEXT            NOT NULL,
			size_bytes      BIGINT          NOT NULL,
			status          TEXT            NOT NULL DEFAULT 'PENDING',
			created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS audit_log (
			id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id     UUID,
			action      TEXT        NOT NULL,
			entity_type TEXT        NOT NULL,
			entity_id   UUID,
			metadata    JSONB,
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_kyc_profiles_user_id   ON kyc_profiles(user_id);
		CREATE INDEX IF NOT EXISTS idx_kyc_documents_profile  ON kyc_do_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_kyc_profiles_user_id   ON kyc_profiles(user_id);
		CREATE INDEX IF NOT EXISTS idx_kyc_documents_profile  ON kyc_doc
