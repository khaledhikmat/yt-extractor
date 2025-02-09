CREATE TABLE jobs (
    id SERIAL PRIMARY KEY,
    channel_id TEXT NOT NULL,
    type TEXT NOT NULL,
    state TEXT NOT NULL,
    videos BIGINT NOT NULL,
    errors BIGINT NOT NULL,
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP
);
