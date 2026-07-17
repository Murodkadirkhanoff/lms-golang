-- Kurs uchun yuklangan thumbnail rasmi (bo'sh = gradient rang ishlatiladi)
ALTER TABLE course.courses
    ADD COLUMN thumbnail_url text NOT NULL DEFAULT '';
