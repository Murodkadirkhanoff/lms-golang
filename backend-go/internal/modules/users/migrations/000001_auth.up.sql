-- users bounded context: owns the `auth` schema.
CREATE SCHEMA IF NOT EXISTS auth;
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS auth.users (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL,
    email citext UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    activated bool NOT NULL DEFAULT false,
    role text NOT NULL DEFAULT 'student'
        CHECK (role IN ('student', 'instructor', 'admin')),
    deleted_at timestamp(0) with time zone DEFAULT NULL,
    version integer NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON auth.users (deleted_at) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS auth.password_reset_tokens (
    hash bytea PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES auth.users ON DELETE CASCADE,
    expiry timestamp(0) with time zone NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_id ON auth.password_reset_tokens (user_id);
