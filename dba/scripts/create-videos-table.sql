CREATE TABLE videos (
    id SERIAL PRIMARY KEY,
    channel_id TEXT NOT NULL,
    video_id TEXT NOT NULL,
    video_url TEXT NOT NULL,
    title TEXT NOT NULL,
    published_at TIMESTAMP NOT NULL,
    views BIGINT NOT NULL,
    comments BIGINT NOT NULL,
    likes BIGINT NOT NULL,
    duration BIGINT NOT NULL,
    short BOOLEAN NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    extraction_url TEXT,
    extracted_at TIMESTAMP,
    externalized_at TIMESTAMP,
    processed_at TIMESTAMP
);
