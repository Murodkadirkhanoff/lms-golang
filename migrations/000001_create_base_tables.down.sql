-- ============================================================
-- LMS Schema — DOWN migration
-- ============================================================

-- Triggerlar
DROP TRIGGER IF EXISTS trg_sync_lesson_access ON enrollments;
DROP TRIGGER IF EXISTS trg_order_total_amount ON order_items;
DROP TRIGGER IF EXISTS trg_category_depth ON categories;

-- Trigger funksiyalar
DROP FUNCTION IF EXISTS sync_lesson_access_on_enrollment();
DROP FUNCTION IF EXISTS sync_order_total_amount();
DROP FUNCTION IF EXISTS set_category_depth();

-- Indexlar
DROP INDEX IF EXISTS idx_categories_deleted_at;
DROP INDEX IF EXISTS idx_courses_deleted_at;
DROP INDEX IF EXISTS idx_users_deleted_at;
DROP INDEX IF EXISTS idx_categories_parent_id;
DROP INDEX IF EXISTS idx_courses_instructor_id;
DROP INDEX IF EXISTS idx_courses_category_id;
DROP INDEX IF EXISTS idx_lesson_access_user_id;
DROP INDEX IF EXISTS idx_enrollments_user_id;
DROP INDEX IF EXISTS idx_order_items_order_id;
DROP INDEX IF EXISTS idx_lessons_module_id;
DROP INDEX IF EXISTS idx_modules_course_id;

-- Jadvallar (foreign key tartibida — avval bog'liqlar)
DROP TABLE IF EXISTS lesson_access;
DROP TABLE IF EXISTS enrollments;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS lessons;
DROP TABLE IF EXISTS modules;
DROP TABLE IF EXISTS courses;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS users;