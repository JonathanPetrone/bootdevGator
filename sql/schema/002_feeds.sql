-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";  

CREATE TABLE feeds (
    id uuid PRIMARY KEY,   
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    name TEXT UNIQUE NOT NULL,
    url TEXT UNIQUE NOT NULL,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE feeds;