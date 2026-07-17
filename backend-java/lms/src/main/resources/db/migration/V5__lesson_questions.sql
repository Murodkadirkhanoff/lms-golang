-- Learn sahifasi Q&A tabi: dars bo'yicha o'quvchi savollari
CREATE TABLE course.lesson_questions (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    lesson_id bigint NOT NULL REFERENCES course.lessons(id) ON DELETE CASCADE,
    -- user auth schema'sida — cross-schema FK yo'q (modul chegarasi), ism snapshot
    user_id bigint NOT NULL,
    user_name text NOT NULL,
    question text NOT NULL
);

CREATE INDEX lesson_questions_lesson_id_idx ON course.lesson_questions (lesson_id);
