-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";  

CREATE TABLE users (
    id uuid PRIMARY KEY,   
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    name TEXT UNIQUE NOT NULL
);

-- +goose Down
DROP TABLE users;