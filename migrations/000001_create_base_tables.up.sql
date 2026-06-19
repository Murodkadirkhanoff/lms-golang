CREATE EXTENSION IF NOT EXISTS citext;

-- users
CREATE TABLE IF NOT EXISTS users (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL,
    email citext UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    activated bool NOT NULL DEFAULT false,
    role text NOT NULL DEFAULT 'student'
        CHECK (role IN ('student', 'instructor', 'admin')),  -- [FIX 1] CHECK qo'shildi
    deleted_at timestamp(0) with time zone DEFAULT NULL,     -- [FIX 7] Soft delete
    version integer NOT NULL DEFAULT 1
);

-- categories (3 tilda nom to'g'ridan-to'g'ri ushbu jadvalda)
CREATE TABLE IF NOT EXISTS categories (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    slug text UNIQUE NOT NULL,
    name_uz text NOT NULL,
    name_ru text NOT NULL,
    name_en text NOT NULL,
    parent_id bigint REFERENCES categories ON DELETE SET NULL,
    depth integer NOT NULL DEFAULT 0,                        -- [FIX 6] Chuqurlikni saqlash
    deleted_at timestamp(0) with time zone DEFAULT NULL,     -- [FIX 7] Soft delete
    version integer NOT NULL DEFAULT 1,

    CONSTRAINT max_category_depth CHECK (depth <= 2)         -- [FIX 6] Max 2 daraja
);

-- categories depth trigger                                  -- [FIX 6]
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

-- courses
CREATE TABLE IF NOT EXISTS courses (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    title text NOT NULL,
    slug text UNIQUE NOT NULL,
    description text NOT NULL DEFAULT '',
    instructor_id bigint NOT NULL REFERENCES users ON DELETE RESTRICT,
    category_id bigint REFERENCES categories ON DELETE SET NULL,
    lang text NOT NULL DEFAULT 'uz'
        CHECK (lang IN ('uz', 'ru', 'en')),
    price numeric(10,2) NOT NULL DEFAULT 0,
    is_published bool NOT NULL DEFAULT false,
    deleted_at timestamp(0) with time zone DEFAULT NULL,     -- [FIX 7] Soft delete
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

-- lessons (module ichidagi darslar)
CREATE TABLE IF NOT EXISTS lessons (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    module_id bigint NOT NULL REFERENCES modules ON DELETE CASCADE,
    title text NOT NULL,
    content_url text NOT NULL DEFAULT '',
    duration_seconds integer NOT NULL DEFAULT 0,
    position integer NOT NULL DEFAULT 0,
    price numeric(10,2) NOT NULL DEFAULT 0,
    is_free bool NOT NULL DEFAULT false,
    version integer NOT NULL DEFAULT 1,

    CONSTRAINT free_lesson_price_check                       -- [FIX 5] is_free + price ziddiyat
        CHECK (is_free = false OR price = 0)
);

-- orders
CREATE TABLE IF NOT EXISTS orders (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    total_amount numeric(10,2) NOT NULL DEFAULT 0,           -- trigger orqali yangilanadi
    status text NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'paid', 'failed', 'cancelled')),  -- [FIX 4]
    version integer NOT NULL DEFAULT 1
);

-- order_items
CREATE TABLE IF NOT EXISTS order_items (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    order_id bigint NOT NULL REFERENCES orders ON DELETE CASCADE,
    course_id bigint REFERENCES courses ON DELETE RESTRICT,
    lesson_id bigint REFERENCES lessons ON DELETE RESTRICT,
    price numeric(10,2) NOT NULL,

    CONSTRAINT order_item_target_check CHECK (
        (course_id IS NOT NULL AND lesson_id IS NULL) OR
        (course_id IS NULL AND lesson_id IS NOT NULL)
    )
);

-- [FIX 1] orders.total_amount avtomatik hisoblash triggeri
CREATE OR REPLACE FUNCTION sync_order_total_amount()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE orders
    SET total_amount = (
        SELECT COALESCE(SUM(price), 0)
        FROM order_items
        WHERE order_id = COALESCE(NEW.order_id, OLD.order_id)
    )
    WHERE id = COALESCE(NEW.order_id, OLD.order_id);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_order_total_amount
    AFTER INSERT OR UPDATE OR DELETE ON order_items
    FOR EACH ROW EXECUTE FUNCTION sync_order_total_amount();

-- enrollments
CREATE TABLE IF NOT EXISTS enrollments (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    course_id bigint NOT NULL REFERENCES courses ON DELETE CASCADE,
    UNIQUE (user_id, course_id)
);

-- lesson_access
CREATE TABLE IF NOT EXISTS lesson_access (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    lesson_id bigint NOT NULL REFERENCES lessons ON DELETE CASCADE,
    UNIQUE (user_id, lesson_id)
);

-- [FIX 2] Kurs sotib olinganda barcha darslar lesson_access ga yoziladi
-- Bu enrollments va lesson_access o'rtasidagi mantiqni birlashtiradi
CREATE OR REPLACE FUNCTION sync_lesson_access_on_enrollment()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO lesson_access (user_id, lesson_id)
    SELECT NEW.user_id, l.id
    FROM lessons l
    JOIN modules m ON m.id = l.module_id
    WHERE m.course_id = NEW.course_id
    ON CONFLICT (user_id, lesson_id) DO NOTHING;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_sync_lesson_access
    AFTER INSERT ON enrollments
    FOR EACH ROW EXECUTE FUNCTION sync_lesson_access_on_enrollment();

-- ============================================================
-- Indexlar
-- ============================================================
CREATE INDEX IF NOT EXISTS idx_modules_course_id ON modules (course_id);
CREATE INDEX IF NOT EXISTS idx_lessons_module_id ON lessons (module_id);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items (order_id);
CREATE INDEX IF NOT EXISTS idx_enrollments_user_id ON enrollments (user_id);
CREATE INDEX IF NOT EXISTS idx_lesson_access_user_id ON lesson_access (user_id);
CREATE INDEX IF NOT EXISTS idx_courses_category_id ON courses (category_id);
CREATE INDEX IF NOT EXISTS idx_courses_instructor_id ON courses (instructor_id);
CREATE INDEX IF NOT EXISTS idx_categories_parent_id ON categories (parent_id);

-- Soft delete uchun indexlar
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_courses_deleted_at ON courses (deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_categories_deleted_at ON categories (deleted_at) WHERE deleted_at IS NULL;