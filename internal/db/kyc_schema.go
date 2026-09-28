package db

import "database/sql"

// MigrateKYC جداول هویت، پرونده KYC، مدارک و audit را می‌سازد.
func MigrateKYC(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			entity_type VARCHAR(16) NOT NULL, -- CORPORATE | BANK
			role VARCHAR(16) NOT NULL DEFAULT 'member', -- member | admin
			status VARCHAR(24) NOT NULL DEFAULT 'PENDING',
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE TABLE IF NOT EXISTS kyc_profiles (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL UNIQUE REFERENCES users(id),
			legal_name VARCHAR(255) NOT NULL,
			registration_number VARCHAR(64),
			tax_id VARCHAR(64),
			jurisdiction VARCHAR(8) NOT NULL,
			contact_phone VARCHAR(32),
			risk_tier VARCHAR(8) NOT NULL DEFAULT 'MEDIUM',
			status VARCHAR(24) NOT NULL DEFAULT 'SUBMITTED', -- SUBMITTED | UNDER_REVIEW | APPROVED | REJECTED
			reviewer_notes TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE TABLE IF NOT EXISTS kyc_documents (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			profile_id UUID NOT NULL REFERENCES kyc_profiles(id) ON DELETE CASCADE,
			document_type VARCHAR(48) NOT NULL,
			file_path VARCHAR(512) NOT NULL,
			original_name VARCHAR(255) NOT NULL,
			checksum_sha256 VARCHAR(64) NOT NULL,
			size_bytes BIGINT NOT NULL,
			status VARCHAR(24) NOT NULL DEFAULT 'SUBMITTED',
			reviewer_notes TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			actor_id UUID,
			actor_email VARCHAR(255),
			action VARCHAR(64) NOT NULL,
			target_entity VARCHAR(64),
			target_id VARCHAR(64),
			ip_address VARCHAR(64),
			metadata JSONB,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_kyc_docs_profile ON kyc_documents(profile_id)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_logs(created_at)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
