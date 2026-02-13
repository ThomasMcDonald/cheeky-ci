-- +goose Up
DROP TABLE IF EXISTS jobs;
CREATE TABLE jobs(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL
);

DROP TABLE IF EXISTS registration_tokens;
CREATE TABLE registration_tokens (
  token_hash TEXT PRIMARY KEY,
  expires_at TIMESTAMP,
  used_at TIMESTAMP NULL
);

DROP TABLE IF EXISTS runners;
CREATE TABLE runners (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(50) NOT NULL,
  token_hash TEXT NOT NULL,
  capabilities JSONB DEFAULT '{}'::jsonb,
  capacity INT DEFAULT 0,
  created_at TIMESTAMP DEFAULT NOW()
);


-- +goose down
--
DROP TABLE IF EXISTS runners;
DROP TABLE IF EXISTS registration_tokens;
DROP TABLE IF EXISTS jobs;
