-- Quiz urinishlari: foydalanuvchi natijalari tarixi (quiz sahifasi "score
-- history") va instruktor analitikasi (o'rtacha ball) uchun.
-- user_id lms_auth'dagi userga ishora qiladi — FK YO'Q (servislararo).
CREATE TABLE IF NOT EXISTS quiz_attempts (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    user_id bigint NOT NULL,
    course_id bigint NOT NULL REFERENCES courses ON DELETE CASCADE,
    score integer NOT NULL CHECK (score BETWEEN 0 AND 100)
);

CREATE INDEX IF NOT EXISTS idx_quiz_attempts_user_course ON quiz_attempts (user_id, course_id);
CREATE INDEX IF NOT EXISTS idx_quiz_attempts_course_id ON quiz_attempts (course_id);
