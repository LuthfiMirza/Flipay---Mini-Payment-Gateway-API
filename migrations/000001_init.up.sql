CREATE TABLE users (
    id UUID PRIMARY KEY,
    name VARCHAR(120) NOT NULL,
    email VARCHAR(160) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE payments (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    reference_no VARCHAR(40) NOT NULL UNIQUE,
    amount BIGINT NOT NULL CHECK (amount > 0),
    payment_method VARCHAR(40) NOT NULL,
    va_number VARCHAR(40) NOT NULL DEFAULT '',
    qris_string TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    expired_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE callbacks (
    id BIGSERIAL PRIMARY KEY,
    payment_id UUID NOT NULL REFERENCES payments(id),
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL,
    attempts INT NOT NULL DEFAULT 1,
    last_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE idempotencies (
    id BIGSERIAL PRIMARY KEY,
    idempotency_key VARCHAR(160) NOT NULL UNIQUE,
    request_hash VARCHAR(128) NOT NULL,
    response JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_payments_user_id ON payments(user_id);
CREATE INDEX idx_payments_status_expired_at ON payments(status, expired_at);
