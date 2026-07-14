-- categories (3 tilda nom to'g'ridan-to'g'ri ushbu jadvalda)
CREATE TABLE IF NOT EXISTS categories (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    slug text UNIQUE NOT NULL,
    name_uz text NOT NULL,
    name_ru text NOT NULL,
    name_en text NOT NULL,
    parent_id bigint REFERENCES categories ON DELETE SET NULL,
    depth integer NOT NULL DEFAULT 0,
    deleted_at timestamp(0) with time zone DEFAULT NULL,
    version integer NOT NULL DEFAULT 1,

    CONSTRAINT max_category_depth CHECK (depth <= 2)
);

CREATE OR REPLACE FUNCTION set_category_depth()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.parent_id IS NULL THEN
        NEW.depth := 0;
    ELSE
        SELECT depth + 1 INTO NEW.depth
        FROM categories WHERE id = NEW.parent_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_category_depth
    BEFORE INSERT OR UPDATE OF parent_id ON categories
    FOR EACH ROW EXECUTE FUNCTION set_category_depth();

-- courses. instructor_id boshqa servis (lms_auth) foydalanuvchisiga ishora
-- qiladi, shuning uchun FK YO'Q — yaxlitlik servis kodida ta'minlanadi.
CREATE TABLE IF NOT EXISTS courses (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    title text NOT NULL,
    slug text UNIQUE NOT NULL,
    description text NOT NULL DEFAULT '',
    instructor_id bigint NOT NULL,
    category_id bigint REFERENCES categories ON DELETE SET NULL,
    lang text NOT NULL DEFAULT 'uz'
        CHECK (lang IN ('uz', 'ru', 'en')),
    price numeric(10,2) NOT NULL DEFAULT 0,
    is_published bool NOT NULL DEFAULT false,
    deleted_at timestamp(0) with time zone DEFAULT NULL,
    version integer NOT NULL DEFAULT 1
);

-- modules (course ichidagi bo'limlar)
CREATE TABLE IF NOT EXISTS modules (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    course_id bigint NOT NULL REFERENCES courses ON DELETE CASCADE,
    title text NOT NULL,
    position integer NOT NULL DEFAULT 0,
    version integer NOT NULL DEFAULT 1
);

-- lessons. type/content frontend talabi: "video" contentUrl ishlatadi,
-- "text" markdown matnni content ustunida saqlaydi.
CREATE TABLE IF NOT EXISTS lessons (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    module_id bigint NOT NULL REFERENCES modules ON DELETE CASCADE,
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

CREATE INDEX IF NOT EXISTS idx_modules_course_id ON modules (course_id);
CREATE INDEX IF NOT EXISTS idx_lessons_module_id ON lessons (module_id);
CREATE INDEX IF NOT EXISTS idx_courses_category_id ON courses (category_id);
CREATE INDEX IF NOT EXISTS idx_courses_instructor_id ON courses (instructor_id);
CREATE INDEX IF NOT EXISTS idx_categories_parent_id ON categories (parent_id);
CREATE INDEX IF NOT EXISTS idx_courses_deleted_at ON courses (deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_categories_deleted_at ON categories (deleted_at) WHERE deleted_at IS NULL;
