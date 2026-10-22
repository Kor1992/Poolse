CREATE TABLE IF NOT EXISTS websites (
    id BIGSERIAL PRIMARY KEY,
    url TEXT NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at	TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE TABLE IF NOT EXISTS checks (
    id BIGSERIAL PRIMARY KEY,
    website_id BIGINT NOT NULL REFERENCES websites(id) ON DELETE CASCADE,
    status_code INTEGER, 
    duration_ms BIGINT NOT NULL CHECK (duration_ms >= 0),
    is_up BOOLEAN NOT NULL,
    error TEXT, 
    checked_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_checks_website_checked_at
ON checks (website_id, checked_at DESC);