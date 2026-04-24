-- +goose Up
CREATE TABLE users (
    id       TEXT PRIMARY KEY,
    username TEXT NOT NULL UNIQUE
);

CREATE TABLE sources (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    title       TEXT NOT NULL,
    link        TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL
);

CREATE TABLE subscriptions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     TEXT NOT NULL,
    source_link TEXT NOT NULL,
    UNIQUE(user_id, source_link)
);

CREATE TABLE credentials (
    user_id TEXT PRIMARY KEY,
    hash    BLOB NOT NULL,
    salt    BLOB NOT NULL
);

CREATE TABLE sessions (
    token      BLOB PRIMARY KEY,
    user_id    TEXT NOT NULL,
    ip         TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    expires_at INTEGER NOT NULL
);

-- +goose Down
DROP TABLE sessions;
DROP TABLE credentials;
DROP TABLE subscriptions;
DROP TABLE sources;
DROP TABLE users;
