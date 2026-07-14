-- course_title — sertifikat berilgan paytdagi snapshot.
CREATE TABLE IF NOT EXISTS certificates (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    issued_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    user_id bigint NOT NULL,
    course_id bigint NOT NULL,
    course_title text NOT NULL DEFAULT '',

    UNIQUE (user_id, course_id)
);

CREATE TABLE IF NOT EXISTS notifications (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    user_id bigint NOT NULL,
    type text NOT NULL DEFAULT 'system'
        CHECK (type IN ('course', 'message', 'promo', 'system')),
    title text NOT NULL,
    body text NOT NULL DEFAULT '',
    read bool NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_certificates_user_id ON certificates (user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications (user_id);
