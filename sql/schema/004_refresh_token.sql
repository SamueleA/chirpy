-- +goose Up
CREATE TABLE tokens (
  token TEXT NOT NULL PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  user_id UUID NOT NULL REFERENCES users ON DELETE CASCADE,
  expires_at TIMESTAMP NOT NULL,
  revoked_at TIMESTAMP
);

--  +goose Down
DROP TABLE tokens;