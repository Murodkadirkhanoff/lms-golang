# 6. `enrollment` modul — yozilish, xarid, progress

```
enrollment/
├── api/
│   └── EnrollmentApi.java        # facade: revenue, enrollmentCounts, accessibleLessonIds
├── EnrollmentDto.java            # {id, createdAt, userId, courseId}
├── OrderDto.java, OrderItemDto.java
├── EnrollmentRepository.java     # enrollments + lesson_access
├── OrderRepository.java          # orders + order_items (+ revenue hisobotlari)
├── CertificateRepository.java    # (ichida CertificateDto)
├── NotificationRepository.java   # (ichida NotificationDto)
├── EnrollmentService.java        # umumiy oqimlar: kirish berish, sertifikat, notify
├── EnrollmentApiImpl.java
├── EnrollmentController.java     # enroll + progress
├── MeController.java             # /v1/me/* (stats, courses, orders, checkout...)
└── TeachingController.java       # /v1/me/teaching/stats
```

DB: `enrollment` schema — `enrollments`, `lesson_access`, `orders`,
`order_items`, `certificates`, `notifications`.

Modulning o'ziga xosligi: `user_id`, `course_id`, `lesson_id` ustunlari
boshqa modullarning ma'lumotiga **FK'siz** ishora qiladi. Kurs/dars
tafsilotlari kerak bo'lganda `CourseApi` facade chaqiriladi.

## 6.1. Ikki asosiy jadval farqi

- **`enrollments`** — "user kursga yozilgan" fakti (kursga a'zolik).
- **`lesson_access`** — "user shu darsga kira oladi" (+ `completed_at` —
  progress belgisi). Kurs sotib olinganda kursning BARCHA darslariga
  yozuv ochiladi; alohida dars sotib olinganda faqat bitta.

Nega ikkalasi ham `course_id` saqlaydi? Progress hisoblashda course
modulga murojaatsiz `GROUP BY course_id` qilish uchun (Go'dagi qaror —
denormalizatsiya).

## 6.2. EnrollmentController

### POST /v1/courses/{id}/enroll — faqat BEPUL kurslar

```java
List<CourseDto> courses = courseApi.coursesByIds(List.of(id));   // kurs facade'dan
...
v.check(course.isPublished, "course", "course is not published");
v.check(course.price == 0, "course", "this course is not free, please purchase it via checkout");
```

Pullik kurs enroll qilinmaydi — checkout orqali sotib olinadi (narx
tekshiruvi server tomonda!).

Idempotent INSERT (`EnrollmentRepository.insert`):

```sql
INSERT INTO enrollment.enrollments (user_id, course_id)
VALUES (:userId, :courseId)
ON CONFLICT (user_id, course_id) DO NOTHING
RETURNING id, created_at
```

Qator qaytsa — yangi yozildi (`isNew=true`); qaytmasa — allaqachon bor,
mavjudi SELECT qilinadi. Ikki marta enroll bosilsa xato chiqmaydi.
`isNew` bo'lsagina: `markEnrolled` (kursning `student_count`++ —
CourseApi orqali) va bildirishnoma yoziladi.

Keyin `grantCourseAccess`: `CourseApi.lessonsForCourse(id)` dan dars
ro'yxati olinib, har biriga `lesson_access` yozuvi (ON CONFLICT DO NOTHING).

### PATCH /v1/enrollments/{id}/progress

```
1. Enrollment topiladi; egasi bo'lmasa 403
2. lesson_access'da completed_at qo'yiladi/olib tashlanadi
   (yozuv yo'q = darsga kirish huquqi yo'q -> 422 "you don't have access to this lesson")
3. completed=true bo'lsa -> maybeIssueCertificate
```

## 6.3. EnrollmentService — yordamchi oqimlar

Uch metodning umumiy printsipi: **asosiy so'rovni yiqitmaydigan yon
ta'sirlar** — xato bo'lsa faqat `log.warn` (Go'dagi xatti-harakat):

- `markEnrolled(courseId)` — student_count++ (statistika, kritik emas);
- `notify(userId, type, title, body)` — bildirishnoma yozish;
- `maybeIssueCertificate(userId, courseId)` — kursning **barcha** darslari
  tugatilganmi tekshiradi:

```java
if (completedCounts.get(courseId) >= course.totalLessons) {
    boolean issued = certificates.issue(...);   // idempotent (ON CONFLICT DO NOTHING)
    if (issued) notify(userId, "course", "Certificate earned", ...);
}
```

Sertifikat `course_title` **snapshot**ini saqlaydi — kurs keyin o'zgarsa
ham sertifikatda o'sha paytdagi nom qoladi.

## 6.4. MeController — checkout (eng murakkab handler)

`POST /v1/me/orders`. Qadamlari:

```
1. Items validatsiyasi: har item AYNAN bitta course_id YOKI lesson_id
   (takrorlar bir marta hisoblanadi — Set bilan dedupe)
2. Kurs/dars ma'lumotlari CourseApi'dan (narxlar DB'dan — clientdan kelgan
   narxga ISHONILMAYDI)
3. Uch himoya tekshiruvi (har biri validatsiya xatosi sifatida):
   - kurs mavjud emas / published emas -> "course does not exist"
   - o'z kursini sotib olish -> "you cannot purchase your own course"
     (dars uchun ham: darsning kursi kimniki ekani tekshiriladi)
   - allaqachon egalik -> "you already own this course/lesson"
4. Order + items INSERT (tranzaksiyada). Status darhol 'paid' — MOCK to'lov.
   total_amount'ni DB trigger hisoblaydi.
5. Kirish berish: kurslar uchun enroll+grantCourseAccess (yangi bo'lsa
   student_count++), darslar uchun bitta lesson_access.
6. "Purchase successful" bildirishnomasi.
7. 201 {"order": {...}}
```

`OrderDto`da JSON uchun alohida "ko'rinish maydonlari" bor
(`finalizeView()`): `id` — satr ko'rinishidagi db id, `date` —
"yyyy-MM-dd", `status` mapping: `paid→completed`,
`failed/cancelled→refunded`, boshqa→`pending`. DB maydonlari
(`dbId`, `createdAt`, `status`) `@JsonIgnore` bilan yashirilgan —
frontend `Order` tipi aynan shu shaklni kutadi.

Order item'larda ham snapshot printsipi: `title`, `instructor_name`,
`thumbnail_color` xarid paytidagi holatida saqlanadi — buyurtmalar
tarixi kurs o'zgarsa/o'chirilsa ham to'g'ri ko'rinadi.

### GET /v1/me/courses — EnrolledCourse

Frontendning learn sahifasi uchun eng "to'la" javob:

```java
record EnrolledCourse(long enrollmentId, CourseDto course, int progress,
                      String currentLesson, int lessonsCompleted,
                      List<Long> completedLessonIds) {}
```

- `course` — CourseApi'dan kelgan **to'liq** CourseDto (Go'da raw JSON
  uzatilardi);
- `progress` — tugatilgan / jami * 100;
- `currentLesson` — tartibdagi **birinchi tugatilmagan** dars nomi
  (darslar tartibi CourseApi.lessonsForCourses'dan, tugatilganlar
  lesson_access'dan);
- kurs o'chirilgan bo'lsa element ro'yxatdan shunchaki tushib qoladi.

### GET /v1/me/stats

Dashboard kartalari: `{enrolled, inProgress, completed, certificates}`.
Kurs "completed" = tugatilgan darslar >= totalLessons; "inProgress" =
kamida bitta dars tugatilgan.

## 6.5. TeachingController — instruktor analitikasi

`GET /v1/me/teaching/stats` — studio dashboard. Ma'lumot ikki moduldan
yig'iladi:

- CourseApi: instruktorning barcha kurslari (draft ham), darslar ro'yxati,
  o'rtacha quiz bali;
- lokal: daromad (`orders`), studentlar soni (`enrollments`), tugatishlar
  (`lesson_access`).

Daromad hisobotida oylik taqsimot — oxirgi 6 oy, bo'sh oylar 0 bilan:

```sql
SELECT to_char(date_trunc('month', o.created_at), 'YYYY-MM') AS month,
       COALESCE(sum(oi.price), 0)
FROM enrollment.order_items oi
JOIN enrollment.orders o ON o.id = oi.order_id
WHERE o.status = 'paid'
  AND (oi.course_id IN (:courseIds) OR oi.lesson_id IN (:lessonIds))
GROUP BY month
```

Har kurs bo'yicha "completion %" = tugatilgan dars yozuvlari /
(studentlar × darslar soni) — engagement jadvali uchun.

## 6.6. EnrollmentApi — facade

```java
public interface EnrollmentApi {
    double revenue();                                        // admin stats uchun
    Map<Long, Integer> enrollmentCountsByUser(List<Long>);   // admin users uchun
    List<Long> accessibleLessonIds(long userId, long courseId); // course paywall uchun
}
```

E'tibor bering: course moduli enrollment'ga (paywall), enrollment couse'ga
(kurs ma'lumoti) bog'liq — **o'zaro** bog'liqlik. Interfeyslar tufayli bu
Spring'da muammosiz (implementatsiyalar emas, abstraksiyalar bog'langan).
Go'da ham xuddi shunday edi (ikkala servis bir-birini HTTP orqali chaqirardi).

Keyingi bo'lim: [database qatlami](07-database.md).
