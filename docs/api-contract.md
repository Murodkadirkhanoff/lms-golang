# LMS API Contract

Bu hujjat frontend (`frontend/src/services/*`) kutayotgan API bilan mavjud Go
backend (`cmd/api`, `internal/data`) o'rtasidagi kelishuvni belgilaydi. Backendni
yozishda shu hujjatga amal qil; frontend allaqachon shu shakllarni kutadi.

## Umumiy qoidalar

- **Base URL:** `/v1` (frontend `NEXT_PUBLIC_API_URL=http://localhost:4000/v1`).
- **Router:** `httprouter` (Greenlight uslubi, mavjud `cmd/api/routes.go`).
- **Response envelope:** har bir resurs nomi bilan o'raladi —
  `{"category": {...}}`, `{"course": {...}}`, `{"user": {...}}`.
- **Ro'yxat (list) javoblari:** frontend `Paginated<T>` kutadi:
  ```json
  { "items": [...], "page": 1, "pageSize": 8, "total": 42 }
  ```
  Diqqat: kalitlar **camelCase** (`pageSize`) — frontend `types/index.ts:105`.
  Resurs envelope (`{"courses": ...}`) emas, balki to'g'ridan-to'g'ri shu obyekt.
- **Xato envelope:** `{"error": ...}` bo'lib, `error` string yoki validatsiya
  uchun `{field: message}` map. Frontend buni `lib/axios.ts:21` da ochadi.
- **JSON kalitlari:** request/DB tomonida **snake_case** (`category_id`,
  `is_published`, `duration_seconds`). Frontend create-da snake_case yuboradi
  (`courses.service.ts:140`).
- **Auth:** `Authorization: Bearer <token>`. Frontend tokenni `localStorage`'da
  saqlaydi va har so'rovga qo'shadi (`lib/axios.ts:9`).
- **Til/role:** `users.role ∈ {student, instructor, admin}`, `lang ∈ {uz, ru, en}`.

---

## 1. Auth & Users

### POST `/v1/users` — Register
Request:
```json
{ "name": "Amir", "email": "a@b.uz", "password": "secret123" }
```
Response `201`: frontend `AuthResult` ({user, token}) kutadi
(`auth.service.ts:52`). Greenlight'da odatda faqat `{"user": ...}` qaytadi va
faollashtirish kerak. **Qaror kerak:** ro'yxatdan o'tgach darrov token berasanmi,
yoki email-activation? Frontend hozir token kutadi — soddaroq yo'l: darrov token.
```json
{ "user": { "id": 1, "name": "Amir", "email": "a@b.uz", "role": "student", "createdAt": "..." }, "token": "..." }
```

### POST `/v1/tokens/authentication` — Login
Request: `{ "email": "...", "password": "..." }`
Response `200` (`auth.service.ts:37`):
```json
{ "user": {...}, "token": "..." }
```
> Frontend `data` ni butunligicha `AuthResult` deb oladi, shuning uchun bu
> endpoint envel`e`siz `{user, token}` qaytarishi kerak (boshqalardan farqi).

### POST `/v1/tokens/password-reset` — Forgot password
Request: `{ "email": "..." }` → `200 { "message": "..." }`.

### PUT `/v1/users/password` — Reset password
Request: `{ "password": "...", "token": "..." }` → `200 { "message": "..." }`.

> **User type:** frontend `User` (`types/index.ts:3`) `role` ni hozir
> ishlatmaydi, lekin RBAC uchun javobga `role` qo'shib qo'y — keyin frontend
> guard qo'shamiz.

---

## 2. Categories  ✅ (allaqachon yozilgan)

Mavjud: `POST/GET/PATCH/DELETE /v1/categories/:id`. Yetishmaydi:

### GET `/v1/categories` — List (frontend `categories.service` kutadi)
Response: frontend qulay shaklda `CategoryNode[]` ham ishlatadi, lekin asosiy
`Category[]` yetadi:
```json
{ "categories": [ { "id":1, "slug":"web", "name_uz":"...", "name_ru":"...", "name_en":"...", "parent_id": null } ] }
```
> `courseCount`, `icon`, `color` (frontend `CategoryNode`) — bular UI-only;
> backend bermasa frontend default beradi.

---

## 3. Courses

DB: `courses` + `modules` + `lessons` (migration `000001`). Frontend `Course`
type richroq (`instructor` obyekt, `rating`, `studentCount`, `totalLessons`,
`reviews`) — bularning bir qismi **hisoblanadi** yoki **keyinroq** qo'shiladi.

### GET `/v1/courses` — List
Query params (`CourseQuery`, `types/index.ts:112`):
`search, category, sort(popular|newest|price-asc|price-desc), page, pageSize`.
Qo'shimcha: `ids` (vergulli), `instructorId`.
Response: `Paginated<Course>` (camelCase meta, yuqoridagi format).

### GET `/v1/courses/:idOrSlug` — Detail
Frontend `getBySlug` **`{ "course": {...} }`** kutadi (`courses.service.ts:81`),
shuningdek `getById` ham `data.course`. Bitta handler id yoki slug qabul qilsin.
`course` ichida `modules[].lessons[]` to'liq bo'lsin (detail sahifa uchun).

### POST `/v1/courses` — Create (auth: instructor/admin)
Request (`courses.service.ts:140`, snake_case):
```json
{
  "title": "...", "description": "...", "category_id": 1,
  "lang": "uz", "price": 0, "is_published": false,
  "modules": [
    { "title": "...", "position": 0,
      "lessons": [
        { "title":"...", "content_url":"...", "duration_seconds":120,
          "position":0, "price":0, "is_free":true }
      ] }
  ]
}
```
`instructor_id` — **serverda** authenticated user'dan olinadi (body'da yo'q).
Slug serverda title'dan generatsiya qilinadi (`data.Slugify` allaqachon bor).
Modules/lessons bitta tranzaksiyada yoziladi. Response `201 { "course": {...} }`.

### Yana kerak bo'ladi (frontend qisman kutadi, routes.go izohida rejalashtirilgan):
- `PATCH /v1/courses/:id`, `DELETE /v1/courses/:id`
- `POST /v1/courses/:id/lessons`, `PATCH/DELETE /v1/lessons/:id`

---

## 4. Enrollment & Learn

DB tayyor: `enrollments`, `lesson_access` + sotib olinganda darslarni ochuvchi
trigger (`migration 000001:160`). routes.go rejasi:
- `POST /v1/courses/:id/enroll` → enrollment yaratadi (trigger lesson_access'ni to'ldiradi).
- `GET /v1/users/:id/enrollments` (yoki `/v1/me/courses`, pastga qara).
- `PATCH /v1/enrollments/:id/progress` → progress yangilash.

> **Diqqat:** progress (`EnrolledCourse.progress`, `lessonsCompleted`) uchun
> hozir DB'da lesson-completion jadvali yo'q. `lesson_access`'ga `completed_at`
> ustuni qo'shish yoki yangi `lesson_progress` jadvali kerak — **yangi migration**.

---

## 5. "Me" (joriy foydalanuvchi) — auth talab qilinadi

Frontend `dashboard.service`, `orders.service`, `notifications.service` kutadi:

| Endpoint | Javob | Izoh |
|---|---|---|
| `GET /v1/me/stats` | dashboard statistikasi | yangi, hisoblanadi |
| `GET /v1/me/courses` | `EnrolledCourse[]` | enrollments + progress |
| `GET /v1/me/orders` | `Order[]` | orders + order_items |
| `GET /v1/me/orders/:id` | `Order` | bitta buyurtma |
| `GET /v1/me/certificates` | `Certificate[]` | ⚠️ jadval yo'q |
| `GET /v1/me/notifications` | `Notification[]` | ⚠️ jadval yo'q |
| `POST /v1/me/notifications/read-all` | `{message}` | ⚠️ jadval yo'q |

`Order` shakli `types/index.ts:134`, `Certificate` `:83`, `Notification` `:145`.

### POST `/v1/me/orders` — Checkout (auth)
Frontend savatchasi **aralash** bo'lishi mumkin: to'liq kurs **va** alohida
darslar (`CartEntry`, `features/cart/cart-context.tsx`). Har bir `order_item`'da
`order_item_target_check` bo'yicha **faqat bittasi** — `course_id` YOKI
`lesson_id` — to'ldiriladi (migration `000001:115`). Kutilayotgan payload:
```json
{
  "items": [
    { "course_id": 1 },
    { "lesson_id": 5 }
  ],
  "payment_method": "card"
}
```
Server narxni **DB'dan** oladi (`courses.price` / `lessons.price`) — clientdan
kelgan narxga ishonma. To'lov muvaffaqiyatli bo'lgach:
- kurs item → `enrollments`'ga yoziladi (trigger barcha darslarni
  `lesson_access`'ga qo'shadi);
- dars item → faqat o'sha dars `lesson_access`'ga qo'shiladi.

> **Biznes-mantiq:** to'liq kurs sotib olingach, uning darslari alohida
> sotilmaydi (frontend `addCourse` savatdagi shu kursning dars itemlarini
> avtomatik olib tashlaydi). Backend ham `enrollments` bor bo'lsa `lesson_access`
> dublikat yaratmasligi kerak (trigger `ON CONFLICT DO NOTHING` buni qiladi).

---

## 6. Instructors

Frontend `instructors.service`: `GET /v1/instructors`, `GET /v1/instructors/:id`.
`Instructor` type (`types/index.ts:21`): `name, headline, students, courses,
rating`. DB'da alohida `instructors` jadvali yo'q — bu `users WHERE
role='instructor'` + hisoblangan maydonlar (kurslar soni, jami studentlar).
`headline`, `avatarColor` uchun `users`'ga profil ustunlari kerak bo'lishi mumkin
(**yangi migration**) yoki default.

---

## 7. Quizzes

Frontend `quiz.service`: `GET /v1/quizzes/:id`. `Quiz` type (`types/index.ts:97`):
`title, passingScore, timeLimitMinutes, questions[]{question, options[],
correctIndex}`. ⚠️ DB'da `quizzes`/`quiz_questions` jadvallari **yo'q** — yangi
migration kerak. PRD'da "Take Quiz" bor, shuning uchun rejaga kiritamiz.

> Xavfsizlik: `correctIndex` ni faqat baholashda ishlat; quizni boshlashda
> javoblarni frontendga yuborma (hozircha mock yuboradi).

---

## 8. Admin — auth: admin

Frontend `admin.service`: `GET /v1/admin/stats`, `GET /v1/admin/users`.
Foydalanuvchilarni boshqarish (role o'zgartirish, bloklash) PRD'da bor —
`PATCH /v1/admin/users/:id` keyin qo'shamiz.

---

## Yetishmayotgan jadvallar (yangi migration kerak)

PRD va frontend talab qiladi, lekin `000001`'da yo'q:
1. `quizzes`, `quiz_questions` (+ `quiz_attempts` natija uchun).
2. `certificates` (kurs tugagach beriladi).
3. `notifications`.
4. `reviews` (frontend `Course.reviews`, `rating`, `ratingCount`).
5. Lesson progress: `lesson_access.completed_at` yoki `lesson_progress`.
6. (ixtiyoriy) instructor profil maydonlari (`headline`, `avatar_color`).

## Implementatsiya tartibi (tavsiya)

1. **Auth** (users register/login, JWT, `authenticate` middleware) — hamma
   narsa shunga bog'liq.
2. **Courses** (list/detail/create) — frontendning yadrosi.
3. **Categories list** (oson, model bor).
4. **Enroll + Me/courses + progress**.
5. **Orders/checkout**.
6. **Quizzes, Certificates, Notifications, Reviews** (yangi migrationlar bilan).
7. **Admin**.

Har bosqichda frontend `.env`'ida `NEXT_PUBLIC_USE_MOCK=false` qilib o'sha
servisni real API'ga ulab tekshiramiz.
