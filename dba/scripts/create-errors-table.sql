CREATE TABLE errors (
    id SERIAL PRIMARY KEY,
    occurred_at TIMESTAMP NOT NULL,
    source TEXT NOT NULL,
    body TEXT NOT NULL
);
