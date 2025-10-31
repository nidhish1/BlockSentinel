-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE IF NOT EXISTS transactions (
    id               BIGSERIAL PRIMARY KEY,
    hash             TEXT UNIQUE NOT NULL,
    from_address     TEXT NOT NULL,
    to_address       TEXT,
    value_wei        NUMERIC(78,0) NOT NULL,
    gas_used         BIGINT,
    gas_price_wei    NUMERIC(78,0),
    block_num        BIGINT NOT NULL,
    block_timestamp  BIGINT NOT NULL,
    input_hex        TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transactions_from ON transactions(from_address);
CREATE INDEX IF NOT EXISTS idx_transactions_to ON transactions(to_address);
CREATE INDEX IF NOT EXISTS idx_transactions_block ON transactions(block_num);

CREATE TABLE IF NOT EXISTS addresses (
    address       TEXT PRIMARY KEY,
    first_seen    TIMESTAMPTZ,
    last_seen     TIMESTAMPTZ,
    labels        TEXT[],
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE IF EXISTS addresses;
DROP INDEX IF EXISTS idx_transactions_block;
DROP INDEX IF EXISTS idx_transactions_to;
DROP INDEX IF EXISTS idx_transactions_from;
DROP TABLE IF EXISTS transactions;


