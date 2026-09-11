CREATE TABLE IF NOT EXISTS sessions(
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,    
    access_token_hash varchar(255) NOT NULL,
    refresh_token_hash varchar(255) NOT NULL,
    access_expires_at TIMESTAMPTZ NOT NULL,
    access_created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    refresh_expires_at TIMESTAMPTZ NOT NULL,
    refresh_created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)