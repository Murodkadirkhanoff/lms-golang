# 8. So'rov hayoti — ikki to'liq misol

Nazariyani birlashtiramiz: ikkita real so'rovning kod bo'ylab to'liq yo'li.

## 8.1. Oddiy misol: POST /v1/users (register)

Frontend yuboradi:

```http
POST /v1/users
Content-Type: application/json

{"name": "Ali Valiyev", "email": "ali@mail.uz", "password": "parol12345"}
```

**1. Security filter chain.** `Authorization` header yo'q →
`JwtAuthFilter` so'rovni anonim o'tkazadi. `SecurityConfig`da `/v1/users`
uchun maxsus qoida yo'q → `anyRequest().permitAll()` — ruxsat.

**2. Routing.** `DispatcherServlet` `POST /v1/users`ni
`AuthController.register`ga bog'laydi (`@RequestMapping("/v1")` +
`@PostMapping("/users")`).

**3. JSON → DTO.** Jackson body'ni `RegisterRequest(name, email, password)`
recordiga o'giradi. Noma'lum kalit kelsa (`fail-on-unknown-properties: true`)
→ `HttpMessageNotReadableException` → handler 400 qaytaradi.

**4. Controller.**

```java
String name = orEmpty(input.name());          // null himoyasi
...
Validator v = new Validator();
validateName(v, name); validateEmail(v, email); validatePassword(v, password);
v.throwIfInvalid();                            // xato -> 422 (handler ushlaydi)

User user = new User();
user.role = Roles.STUDENT;                     // register doim student
user.passwordHash = passwordEncoder.encode(password).getBytes(UTF_8);  // bcrypt(12)

try {
    users.insert(user);                        // INSERT ... RETURNING
} catch (UserRepository.DuplicateEmailException e) {
    v.addError("email", "a user with this email address already exists");
    v.throwIfInvalid();                        // -> 422
}

String token = jwtService.newToken(user.id, user.role);
return ResponseEntity.status(CREATED).body(Map.of("user", UserResponse.from(user), "token", token));
```

**5. Javob.** Jackson `UserResponse` recordini camelCase JSON'ga o'giradi:

```json
201 Created
{"user": {"id": 5, "createdAt": "2026-07-16T10:31:13Z", "name": "Ali Valiyev",
          "email": "ali@mail.uz", "role": "student"},
 "token": "eyJhbGciOiJIUzI1NiJ9..."}
```

Xato stsenariylar (hammasi `GlobalExceptionHandler` orqali):
- parol 5 belgi → `422 {"error": {"password": "must be at least 8 bytes long"}}`
- email band → `422 {"error": {"email": "a user with this email address already exists"}}`
- buzuq JSON → `400 {"error": "body contains badly-formed JSON"}`

## 8.2. Murakkab misol: POST /v1/me/orders (checkout)

Uch modul qatnashadi: enrollment (egasi), course (narx/darslar),
DB trigger (total). Student 2 ta narsani sotib olmoqda:

```http
POST /v1/me/orders
Authorization: Bearer eyJ...

{"items": [{"course_id": 7}, {"lesson_id": 42}], "payment_method": "card"}
```

**1. Filter chain.** `JwtAuthFilter` tokenni parse qiladi →
`UserPrincipal(id=5, role=student)` SecurityContext'ga tushadi.
`SecurityConfig`: `/v1/me/**` → `authenticated()` ✓.

**2. `MeController.checkout`** (`@JsonNaming(SnakeCase)` tufayli
`course_id` → `courseId`):

```
items validatsiyasi:
  har item aynan bitta course_id YOKI lesson_id  (ikkalasi/hech biri -> 422)
  dedupe: courseIds=[7], lessonIds=[42]
```

**3. Course modulidan ma'lumot (facade chaqiruvlar):**

```java
Map<Long, CourseDto> courses = coursesById(List.of(7L));        // CourseApi.coursesByIds
Map<Long, LessonInfo> lessons = ...courseApi.lessonsByIds([42]) // narx, kurs nomi, courseId
// 42-dars qaysi kursniki? uning egasini bilish uchun o'sha kurs ham olinadi
```

Bu chaqiruvlar ortida `CourseRepository`ning LATERAL join'li SELECT'i +
`decorate()` (instruktor ismlari `UserApi`dan) ishlaydi. Go'da xuddi shu
joyda ikkita HTTP so'rov ketardi (`/internal/courses?ids=`,
`/internal/lessons?ids=`).

**4. Himoya tekshiruvlari** (lokal jadvallar bilan):

```java
Set<Long> ownedCourses = enrollments.ownedCourses(5, [7]);   // allaqachon yozilganmi?
Set<Long> ownedLessons = enrollments.ownedLessons(5, [42]);
// + kurs published'mi, o'z kursi emasmi, narx DB'dan (clientga ishonch yo'q)
```

Birortasi buzilsa — hamma xato yig'ilib bitta 422 javobda:
`{"error": {"items[0].course_id": "you already own this course"}}`.

**5. Yozish** (`OrderRepository.insert`, `@Transactional`):

```sql
INSERT INTO enrollment.orders (user_id, status, payment_method)
VALUES (5, 'paid', 'card') RETURNING id, created_at;          -- mock to'lov: darhol paid

INSERT INTO enrollment.order_items (order_id, course_id, ..., price) VALUES (...);  -- item 1
INSERT INTO enrollment.order_items (order_id, lesson_id, ..., price) VALUES (...);  -- item 2
-- har INSERT'da DB trigger orders.total_amount'ni qayta hisoblaydi
```

**6. Kirish berish:**

```java
for (kurs 7):
    enrollments.insert(5, 7)                       // ON CONFLICT DO NOTHING
    service.grantCourseAccess(5, 7)                // CourseApi.lessonsForCourse -> lesson_access'ga hammasi
    service.markEnrolled(7)                        // CourseApi.incrementStudentCount (student_count++)

for (dars 42):
    enrollments.grantLessonAccess(5, lesson.courseId, [42])   // faqat shu dars
```

**7. Bildirishnoma + javob:**

```json
201 Created
{"order": {"id": "3", "date": "2026-07-16", "status": "completed",
           "paymentMethod": "card", "total": 65000.0,
           "items": [{"courseTitle": "Advanced Java", "instructor": "Test Instructor",
                      "thumbnailColor": "bg-fuchsia-200", "price": 50000.0},
                     {"courseTitle": "Java Basics", "instructor": "",
                      "thumbnailColor": "bg-sky-200", "price": 15000.0}]}}
```

Endi bu user kurs sahifasini ochsa (`GET /v1/courses/7`) paywall
`enrollmentApi.accessibleLessonIds(5, 7)` orqali kirishini ko'radi va
`content`/`contentUrl` yashirilmaydi.

## 8.3. Modullararo bog'liqlik xaritasi

```
            UserApi (auth)
              ▲         ▲
              │         │
        CourseService  ReviewController, InstructorController
        (instruktor     (ism snapshot / instruktor ismi)
         ismlari)
              
   CourseApi (course)                    EnrollmentApi (enrollment)
     ▲    ▲    ▲                            ▲            ▲
     │    │    │                            │            │
 Admin  MeController,                  CourseController  AdminController
(stats, EnrollmentController,          (paywall)         (stats, users)
 counts) TeachingController
        (kurs ma'lumoti, darslar,
         student_count++, quiz stats)
```

Hech bir strelka `api/` paketidan tashqariga kirmaydi — modulni ajratish
kerak bo'lganda aynan shu strelkalar HTTP/gRPC'ga aylanadi.

## 8.4. O'rganishni davom ettirish uchun

Kod o'qish tartibi bo'yicha tavsiya (soddadan murakkabga):

1. `shared/HealthController` → `AuthController` → `UserRepository`
2. `CategoryController`/`CategoryRepository` (to'liq CRUD namunasi)
3. `CourseRepository.list` (dinamik SQL) → `CourseController.show` (paywall)
4. `MeController.checkout` (hamma tushunchalar bir joyda)

Amaliy mashqlar (kod bazasini his qilish uchun):

- Yangi endpoint qo'sh: `GET /v1/me/reviews` — userning barcha sharhlari
  (ReviewRepository'ga metod + MeController'ga route... lekin sharhlar
  course modulida — facade kerakmi yoki endpoint course modulida
  bo'lishi kerakmi? — modul chegarasi haqida o'ylashga majbur qiladi).
- `POST /v1/tokens/password-reset`ga rate-limit qo'shib ko'r
  (filter yozish mashqi).
- Bitta integration test yoz (Testcontainers bilan) — register+login oqimi.
