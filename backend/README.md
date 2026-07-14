# LMS Backend — Microservices

3 ta Go microservice + nginx gateway + PostgreSQL (bitta konteyner, har servisga alohida database).

```
frontend (:3000) ──► nginx gateway (:4000)
                       ├─ /v1/users, /v1/tokens, /v1/admin ──► auth-service (:4001, DB: lms_auth)
                       ├─ /v1/categories, /v1/courses, /v1/lessons,
                       │  /v1/instructors, /v1/quizzes ───────► course-service (:4002, DB: lms_course)
                       └─ /v1/me, /v1/enrollments,
                          /v1/courses/:id/enroll ─────────────► enrollment-service (:4003, DB: lms_enrollment)

enrollment-service ──HTTP──► course-service /internal/* (narx, darslar, student_count)
course-service     ──HTTP──► auth-service /internal/users (instruktor ismlari)
auth-service       ──HTTP──► course/enrollment /internal/stats (admin panel)
```

## Texnologiyalar

- **Router:** Chi v5 · **Auth:** JWT HS256 (umumiy `JWT_SECRET`, har servis o'zi tekshiradi)
- **Servislararo aloqa:** ichki REST `/internal/*` (`X-Internal-Key` header, gateway tashqariga chiqarmaydi)
- **Migratsiyalar:** golang-migrate, har servis startup'da o'zinikini qo'llaydi
- **Umumiy kod:** `pkg/` moduli (go.work workspace): jsonutil, httperr, validator, auth, middleware, database, svcclient, env, uidefaults

## Ishga tushirish

`.env.example` → `.env` nusxalab qiymatlarni to'ldiring.

**Dev rejim** (tez hot-reload — servislar hostda):

```
make up              # db + gateway (docker)
make run/auth        # har biri alohida terminalda
make run/course
make run/enrollment
cd ../frontend && npm run dev
```

**Full rejim** (butun tizim docker'da — prod-simulyatsiya):

```
make up/full         # db + 3 servis + gateway + frontend, hammasi build qilinadi
```

Full rejimda faqat `:4000` (API) va `:3000` (frontend) ochiq; Go servislar,
`/internal/*` endpointlar va Postgres docker network ichida yopiq qoladi.

To'xtatish: `make down`

## API konvensiyalari

- Base URL: `/v1`, javob envelope: `{"course": ...}`, ro'yxatlar: `{"items": [...], "page", "pageSize", "total"}`
- **Javoblar camelCase** (`nameUz`, `isPublished`, `createdAt`), **requestlar snake_case** (`name_uz`, `course_id`) — frontend `types/index.ts` va `services/*` shakllariga mos
- Xato envelope: `{"error": string | {field: message}}`
- Quiz kurs ID bilan so'raladi: `GET /v1/quizzes/{courseId}` (frontend `ROUTES.quiz(course.id)`)
- Kurs yaratishga rol talab qilinmaydi (no-role-gating); RBAC faqat `/v1/admin/*` (admin roli)

## Hozircha mock qolgan

- **To'lov** — checkout'da buyurtma darhol `paid` bo'ladi. Provider (Payme/Click/Stripe)
  ulanganda: `checkoutHandler`da status `pending` + webhook'da `paid`ga o'tkazib,
  access berish bloki webhook'ka ko'chiriladi.
- **Email** — parol tiklash tokeni yaratiladi va ishlaydi, lekin SMTP yo'qligi
  uchun token servis logiga yoziladi (`auth-service` logida `password reset token`).
