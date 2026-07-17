-- course moduli: kategoriyalar, kurslar, modullar, darslar, sharhlar, quizlar
-- (Go: backend/course-service/migrations/000001..000003)

CREATE TABLE IF NOT EXISTS course.categories (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    slug text UNIQUE NOT NULL,
    name_uz text NOT NULL,
    name_ru text NOT NULL,
    name_en text NOT NULL,
    parent_id bigint REFERENCES course.categories ON DELETE SET NULL,
    depth integer NOT NULL DEFAULT 0,
    deleted_at timestamp(0) with time zone DEFAULT NULL,
    version integer NOT NULL DEFAULT 1,

    CONSTRAINT max_category_depth CHECK (depth <= 2)
);

CREATE OR REPLACE FUNCTION course.set_category_depth()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.parent_id IS NULL THEN
        NEW.depth := 0;
    ELSE
        SELECT depth + 1 INTO NEW.depth
        FROM course.categories WHERE id = NEW.parent_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_category_depth
    BEFORE INSERT OR UPDATE OF parent_id ON course.categories
    FOR EACH ROW EXECUTE FUNCTION course.set_category_depth();

-- instructor_id auth modulining useriga ishora qiladi — modul chegarasi
-- sababli FK YO'Q, yaxlitlik kodda ta'minlanadi.
CREATE TABLE IF NOT EXISTS course.courses (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    title text NOT NULL,
    slug text UNIQUE NOT NULL,
    description text NOT NULL DEFAULT '',
    instructor_id bigint NOT NULL,
    category_id bigint REFERENCES course.categories ON DELETE SET NULL,
    lang text NOT NULL DEFAULT 'uz'
        CHECK (lang IN ('uz', 'ru', 'en')),
    price numeric(10,2) NOT NULL DEFAULT 0,
    is_published bool NOT NULL DEFAULT false,
    student_count integer NOT NULL DEFAULT 0,
    deleted_at timestamp(0) with time zone DEFAULT NULL,
    version integer NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS course.modules (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    course_id bigint NOT NULL REFERENCES course.courses ON DELETE CASCADE,
    title text NOT NULL,
    position integer NOT NULL DEFAULT 0,
    version integer NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS course.lessons (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    module_id bigint NOT NULL REFERENCES course.modules ON DELETE CASCADE,
    title text NOT NULL,
    type text NOT NULL DEFAULT 'video'
        CHECK (type IN ('video', 'text')),
    content_url text NOT NULL DEFAULT '',
    content text NOT NULL DEFAULT '',
    duration_seconds integer NOT NULL DEFAULT 0,
    position integer NOT NULL DEFAULT 0,
    price numeric(10,2) NOT NULL DEFAULT 0,
    is_free bool NOT NULL DEFAULT false,
    version integer NOT NULL DEFAULT 1,

    CONSTRAINT free_lesson_price_check
        CHECK (is_free = false OR price = 0)
);

-- user_id/user_name auth useriga ishora (FK yo'q, ism snapshot).
CREATE TABLE IF NOT EXISTS course.reviews (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    course_id bigint NOT NULL REFERENCES course.courses ON DELETE CASCADE,
    user_id bigint NOT NULL,
    user_name text NOT NULL DEFAULT '',
    rating integer NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment text NOT NULL DEFAULT '',

    UNIQUE (course_id, user_id)
);

-- Har kursga bitta quiz (frontend quiz'ni kurs id bilan so'raydi).
CREATE TABLE IF NOT EXISTS course.quizzes (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    course_id bigint UNIQUE NOT NULL REFERENCES course.courses ON DELETE CASCADE,
    title text NOT NULL,
    passing_score integer NOT NULL DEFAULT 70,
    time_limit_minutes integer NOT NULL DEFAULT 10,
    version integer NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS course.quiz_questions (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    quiz_id bigint NOT NULL REFERENCES course.quizzes ON DELETE CASCADE,
    question text NOT NULL,
    options text[] NOT NULL,
    correct_index integer NOT NULL DEFAULT 0,
    position integer NOT NULL DEFAULT 0
);

-- user_id auth useriga ishora — FK YO'Q.
CREATE TABLE IF NOT EXISTS course.quiz_attempts (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    user_id bigint NOT NULL,
    course_id bigint NOT NULL REFERENCES course.courses ON DELETE CASCADE,
    score integer NOT NULL CHECK (score BETWEEN 0 AND 100)
);

CREATE INDEX IF NOT EXISTS idx_modules_course_id ON course.modules (course_id);
CREATE INDEX IF NOT EXISTS idx_lessons_module_id ON course.lessons (module_id);
CREATE INDEX IF NOT EXISTS idx_courses_category_id ON course.courses (category_id);
CREATE INDEX IF NOT EXISTS idx_courses_instructor_id ON course.courses (instructor_id);
CREATE INDEX IF NOT EXISTS idx_categories_parent_id ON course.categories (parent_id);
CREATE INDEX IF NOT EXISTS idx_courses_deleted_at ON course.courses (deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_categories_deleted_at ON course.categories (deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_reviews_course_id ON course.reviews (course_id);
CREATE INDEX IF NOT EXISTS idx_quiz_questions_quiz_id ON course.quiz_questions (quiz_id);
CREATE INDEX IF NOT EXISTS idx_quiz_attempts_user_course ON course.quiz_attempts (user_id, course_id);
CREATE INDEX IF NOT EXISTS idx_quiz_attempts_course_id ON course.quiz_attempts (course_id);
