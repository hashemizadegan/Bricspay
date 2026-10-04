package db

import "database/sql"

func MigrateBankCards(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS bank_cards (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  last4 VARCHAR(4) NOT NULL,
  brand TEXT,
  holder_name TEXT,
  expiry_month SMALLINT,
  expiry_year SMALLINT,
  display_name TEXT,
  token_ref TEXT,
  is_preferred BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT bank_cards_last4_chk CHECK (last4 ~ '^[0-9]{4}$')
);
CREATE INDEX IF NOT EXISTS idx_bank_cards_user ON bank_cards(user_id);

CREATE OR REPLACE FUNCTION enforce_bank_card_limit() RETURNS trigger AS $$
BEGIN
  IF (SELECT COUNT(*) FROM bank_cards WHERE user_id = NEW.user_id) >= 3 THEN
    RAISE EXCEPTION 'card_limit_exceeded' USING ERRCODE = 'P0001';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_bank_card_limit ON bank_cards;
CREATE TRIGGER trg_bank_card_limit
BEFORE INSERT ON bank_cards
FOR EACH ROW EXECUTE FUNCTION enforce_bank_card_limit();

CREATE OR REPLACE FUNCTION enforce_single_preferred_card() RETURNS trigger AS $$
BEGIN
  IF NEW.is_preferred THEN
    UPDATE bank_cards SET is_preferred = FALSE
     WHERE user_id = NEW.user_id AND id <> NEW.id AND is_preferred = TRUE;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_bank_card_preferred ON bank_cards;
CREATE TRIGGER trg_bank_card_preferred
BEFORE INSERT OR UPDATE OF is_preferred ON bank_cards
FOR EACH ROW EXECUTE FUNCTION enforce_single_preferred_card();
`)
	return err
}
