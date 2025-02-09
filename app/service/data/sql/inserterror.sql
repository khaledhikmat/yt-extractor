INSERT INTO errors (
    source, body, occurred_at
) VALUES (
    $1, $2, NOW()
)
RETURNING id
