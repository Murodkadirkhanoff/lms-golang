CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users (
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

CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at) WHERE deleted_at IS NULL;
