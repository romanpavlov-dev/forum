CREATE TABLE IF NOT EXISTS users(
    id BIGSERIAL PRIMARY KEY,
    username varchar(50) NOT NULL,
    email varchar(255) NOT NULL,
    password_hash text NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)

