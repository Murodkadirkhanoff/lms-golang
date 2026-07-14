-- user_id / course_id / lesson_id boshqa servislar (lms_auth, lms_course)
-- ma'lumotlariga ishora qiladi — FK YO'Q, yaxlitlik servis kodida.

CREATE TABLE IF NOT EXISTS enrollments (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    user_id bigint NOT NULL,
    course_id bigint NOT NULL,
    UNIQUE (user_id, course_id)
);

-- course_id shu jadvalda saqlanadi, shunda kurs bo'yicha progress hisoblash
-- uchun course-service'ga murojaat kerak bo'lmaydi. completed_at — dars
-- tugatilgan payt (progress kontrakt §4).
CREATE TABLE IF NOT EXISTS lesson_access (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    user_id bigint NOT NULL,
    lesson_id bigint NOT NULL,
    course_id bigint NOT NULL DEFAULT 0,
    completed_at timestamp(0) with time zone DEFAULT NULL,
    UNIQUE (user_id, lesson_id)
);

CREATE TABLE IF NOT EXISTS orders (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    user_id bigint NOT NULL,
    total_amount numeric(10,2) NOT NULL DEFAULT 0,
    payment_method text NOT NULL DEFAULT 'card',
    status text NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'paid', 'failed', 'cancelled')),
    version integer NOT NULL DEFAULT 1
);

-- title/instructor_name/thumbnail_color — sotib olish paytidagi snapshot
-- (kurs keyin o'zgarsa ham buyurtma tarixi to'g'ri ko'rinadi).
CREATE TABLE IF NOT EXISTS order_items (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    order_id bigint NOT NULL REFERENCES orders ON DELETE CASCADE,
    course_id bigint,
    lesson_id bigint,
    title text NOT NULL DEFAULT '',
    instructor_name text NOT NULL DEFAULT '',
    thumbnail_color text NOT NULL DEFAULT '',
    price numeric(10,2) NOT NULL,

    CONSTRAINT order_item_target_check CHECK (
        (course_id IS NOT NULL AND lesson_id IS NULL) OR
        (course_id IS NULL AND lesson_id IS NOT NULL)
    )
);

-- orders.total_amount avtomatik hisoblash triggeri
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

CREATE INDEX IF NOT EXISTS idx_enrollments_user_id ON enrollments (user_id);
CREATE INDEX IF NOT EXISTS idx_lesson_access_user_id ON lesson_access (user_id);
CREATE INDEX IF NOT EXISTS idx_lesson_access_user_course ON lesson_access (user_id, course_id);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items (order_id);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders (user_id);
