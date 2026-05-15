CREATE TABLE IF NOT EXISTS users (
  user_id BIGINT PRIMARY KEY,
  name    TEXT   NOT NULL,
  status  TEXT   NOT NULL
);
