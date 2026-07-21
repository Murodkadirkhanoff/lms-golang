-- enrollment bounded context: owns the `enrollment` schema.
-- user_id / course_id / lesson_id reference other contexts; no cross-schema FK
-- by design (module boundary).
CREATE SCHEMA IF NOT EXISTS enrollment;

CREATE TABLE IF NOT EXISTS enrollment.enrollments (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    user_id bigint NOT NULL,
    course_id bigint NOT NULL,
    UNIQUE (user_id, course_id)
);

CREATE TABLE IF NOT EXISTS enrollment.lesson_access (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    user_id bigint NOT NULL,
    lesson_id bigint NOT NULL,
    course_id bigint NOT NULL DEFAULT 0,
    completed_at timestamp(0) with time zone DEFAULT NULL,
    UNIQUE (user_id, lesson_id)
);

CREATE TABLE IF NOT EXISTS enrollment.orders (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    user_id bigint NOT NULL,
    total_amount numeric(10,2) NOT NULL DEFAULT 0,
    payment_method text NOT NULL DEFAULT 'card',
    status text NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'paid', 'failed', 'cancelled')),
    version integer NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS enrollment.order_items (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    order_id bigint NOT NULL REFERENCES enrollment.orders ON DELETE CASCADE,
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

CREATE OR REPLACE FUNCTION enrollment.sync_order_total_amount()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE enrollment.orders
    SET total_amount = (
        SELECT COALESCE(SUM(price), 0)
        FROM enrollment.order_items
        WHERE order_id = COALESCE(NEW.order_id, OLD.order_id)
    )
    WHERE id = COALESCE(NEW.order_id, OLD.order_id);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_order_total_amount ON enrollment.order_items;
CREATE TRIGGER trg_order_total_amount
    AFTER INSERT OR UPDATE OR DELETE ON enrollment.order_items
    FOR EACH ROW EXECUTE FUNCTION enrollment.sync_order_total_amount();

CREATE TABLE IF NOT EXISTS enrollment.certificates (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    issued_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    user_id bigint NOT NULL,
    course_id bigint NOT NULL,
    course_title text NOT NULL DEFAULT '',
    UNIQUE (user_id, course_id)
);

CREATE TABLE IF NOT EXISTS enrollment.notifications (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    user_id bigint NOT NULL,
    type text NOT NULL DEFAULT 'system'
        CHECK (type IN ('course', 'message', 'promo', 'system')),
    title text NOT NULL,
    body text NOT NULL DEFAULT '',
    read bool NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_enrollments_user_id ON enrollment.enrollments (user_id);
CREATE INDEX IF NOT EXISTS idx_lesson_access_user_id ON enrollment.lesson_access (user_id);
CREATE INDEX IF NOT EXISTS idx_lesson_access_user_course ON enrollment.lesson_access (user_id, course_id);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON enrollment.order_items (order_id);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON enrollment.orders (user_id);
CREATE INDEX IF NOT EXISTS idx_certificates_user_id ON enrollment.certificates (user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON enrollment.notifications (user_id);
