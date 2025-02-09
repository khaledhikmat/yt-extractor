INSERT INTO api_keys (
    key, started_at, expires_at
) VALUES (
    :key, NOW(), NOW() + INTERVAL '1 year'
)
RETURNING id
