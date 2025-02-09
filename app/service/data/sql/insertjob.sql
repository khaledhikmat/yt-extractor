INSERT INTO jobs (
    channel_id, type, state, videos, errors, started_at, completed_at
) VALUES (
    :channel_id, :type, :state, :videos, :errors, :started_at, :completed_at
)
RETURNING id
