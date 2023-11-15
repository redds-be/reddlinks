-- +goose Up

CREATE TABLE links (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    expire_at TIMESTAMP NOT NULL,
    url TEXT UNIQUE NOT NULL
);

-- +goose Down

DROP TABLE links;