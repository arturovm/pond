-- +goose Up
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

-- +goose Down
DROP TABLE subscriptions;
DROP TABLE sources;
