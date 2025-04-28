
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(20) NOT NULL, 
    last_name VARCHAR(20) NOT NULL, 
    phone_number VARCHAR(20) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    status VARCHAR(16) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)


CREATE TABLE refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)


CREATE UNIQUE INDEX idx_refresh_tokens_token ON refresh_tokens(token);

CREATE TABLE wallets (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    balance NUMERIC(20, 2) NOT NULL DEFAULT 0.0, 
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
)

CREATE UNIQUE INDEX idx_wallets_user_id ON wallets(user_id);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    type VARCHAR(20) NOT NULL CHECK (type IN('deposit', 'withdrawal', 'transfer_debited', 'transfer_credited')), 
    from_wallet_id INTEGER NOT NULL,
    to_wallet_id INTEGER NULL,
    amount NUMERIC(20, 2) NOT NULL CHECK (amount > 0.0), 
    fee NUMERIC(8, 2) NOT NULL DEFAULT 0.0 CHECK(fee >= 0.0), 
    description TEXT,
    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_from_wallet FOREIGN KEY (from_wallet_id) REFERENCES wallets(id) ON DELETE SET NULL
)
