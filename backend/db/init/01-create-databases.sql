-- Har bir microservice uchun alohida database.
-- Bu script faqat YANGI postgres volume'da ishlaydi (docker-entrypoint-initdb.d).
CREATE DATABASE lms_auth;
CREATE DATABASE lms_course;
CREATE DATABASE lms_enrollment;
