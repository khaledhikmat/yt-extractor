CREATE TABLE api_keys (
    id SERIAL PRIMARY KEY,
    key TEXT NOT NULL,
    started_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL
);
