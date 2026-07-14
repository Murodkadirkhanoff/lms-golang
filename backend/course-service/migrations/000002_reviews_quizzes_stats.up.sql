-- user_id/user_name lms_auth foydalanuvchisiga ishora (FK yo'q, snapshot).
CREATE TABLE IF NOT EXISTS reviews (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    course_id bigint NOT NULL REFERENCES courses ON DELETE CASCADE,
    user_id bigint NOT NULL,
    user_name text NOT NULL DEFAULT '',
    rating integer NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment text NOT NULL DEFAULT '',

    UNIQUE (course_id, user_id)
);

-- Har kursga bitta quiz (frontend quiz'ni kurs id bilan so'raydi).
CREATE TABLE IF NOT EXISTS quizzes (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    course_id bigint UNIQUE NOT NULL REFERENCES courses ON DELETE CASCADE,
    title text NOT NULL,
    passing_score integer NOT NULL DEFAULT 70,
    time_limit_minutes integer NOT NULL DEFAULT 10,
    version integer NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS quiz_questions (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    quiz_id bigint NOT NULL REFERENCES quizzes ON DELETE CASCADE,
    question text NOT NULL,
    options text[] NOT NULL,
    correct_index integer NOT NULL DEFAULT 0,
    position integer NOT NULL DEFAULT 0
);

-- Denormalizatsiya: enrollment-service yangi yozilganda internal endpoint
-- orqali oshiradi. Shu tufayli "popular" sort SQL'da ishlaydi.
ALTER TABLE courses ADD COLUMN IF NOT EXISTS student_count integer NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_reviews_course_id ON reviews (course_id);
CREATE INDEX IF NOT EXISTS idx_quiz_questions_quiz_id ON quiz_questions (quiz_id);
