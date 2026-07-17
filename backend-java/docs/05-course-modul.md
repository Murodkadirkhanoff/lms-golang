# 5. `course` modul — kurslar katalogi

Eng katta modul: kategoriyalar, kurslar (modullar/darslar bilan), sharhlar,
quizlar, instruktorlar.

```
course/
├── api/                          # PUBLIC — boshqa modullar va JSON javoblar uchun
│   ├── CourseApi.java            # facade interfeysi (+ LessonInfo, CourseStats recordlari)
│   ├── CourseDto.java            # frontend Course tipi bilan bir xil JSON
│   ├── ModuleDto.java, LessonDto.java, ReviewDto.java, InstructorDto.java
├── CategoryDto.java, QuizDto.java, QuizAttemptDto.java   # ichki DTO'lar
├── Slugs.java                    # slugify util
├── CourseErrors.java             # domen exceptionlari
├── CategoryRepository.java
├── CourseRepository.java         # eng murakkab SQL shu yerda
├── ReviewRepository.java
├── QuizRepository.java
├── CourseService.java            # CourseApi impl + dekoratsiya
├── CategoryController.java
├── CourseController.java
├── QuizController.java
├── ReviewController.java
└── InstructorController.java
```

DB: `course` schema — `categories`, `courses`, `modules`, `lessons`,
`reviews`, `quizzes`, `quiz_questions`, `quiz_attempts`.

## 5.1. Nega DTO'lar `api/` paketida public?

`CourseDto` ikki vazifada: (1) `/v1/courses` javoblarining JSON shakli,
(2) enrollment moduli `me/courses` javobiga kursni **butunligicha**
joylaydi (Go'da course-service JSON'i raw uzatilardi). Shuning uchun u
public va public maydonli:

```java
public class CourseDto {
    public long id;
    public Instant createdAt;
    public String slug, title, description, thumbnailColor;
    public Long categoryId;        // null bo'lishi mumkin -> JSON'da null
    public String category = "";   // kategoriya slug'i
    ...
    @JsonInclude(JsonInclude.Include.NON_NULL)
    public List<ModuleDto> modules;   // ro'yxatda null (JSON'da chiqmaydi), detalda bor
    @JsonInclude(JsonInclude.Include.NON_NULL)
    public List<ReviewDto> reviews;

    @JsonIgnore
    public long instructorId;      // JSON'ga chiqmaydi (instructor obyekti chiqadi)
    @JsonIgnore
    public int version;
}
```

`@JsonInclude(NON_NULL)` — Go'dagi `omitempty`ning ekvivalenti: ro'yxat
so'rovida modules yuklanmaydi (null) va JSON'da ham chiqmaydi; detal
so'rovida to'ldiriladi.

## 5.2. CourseRepository — murakkab SQL'ni qanday ko'chirdik

Go'dagi so'rovlar deyarli 1:1 (faqat `$1` o'rniga `:nomli` parametrlar va
`course.` schema prefiksi). Asosiy `LIST_SELECT`:

```sql
SELECT count(*) OVER() AS total, c.id, ..., 
       COALESCE(agg.total_lessons, 0), COALESCE(agg.total_seconds, 0),
       COALESCE(rv.avg_rating, 0), COALESCE(rv.rating_count, 0)
FROM course.courses c
LEFT JOIN course.categories cat ON cat.id = c.category_id AND cat.deleted_at IS NULL
LEFT JOIN LATERAL (           -- har kurs uchun darslar soni va davomiyligi
    SELECT count(l.id) AS total_lessons, COALESCE(sum(l.duration_seconds), 0) AS total_seconds
    FROM course.modules m JOIN course.lessons l ON l.module_id = m.id
    WHERE m.course_id = c.id
) agg ON true
LEFT JOIN LATERAL (           -- har kurs uchun o'rtacha reyting
    SELECT round(avg(r.rating)::numeric, 1) AS avg_rating, count(r.id) AS rating_count
    FROM course.reviews r WHERE r.course_id = c.id
) rv ON true
```

`LATERAL` — Postgres'ga xos "har qator uchun subquery". Eloquent
`withCount`/`withAvg` bir necha so'rovga bo'lib bajargan ishni bitta
so'rovda qiladi.

Filtrlash dinamik quriladi (Query Builder'siz, oddiy StringBuilder):

```java
if (!filters.ids().isEmpty())        sql.append(" AND c.id IN (:ids)");
if (filters.instructorId() != 0)     sql.append(" AND c.instructor_id = :instructorId");
if (!filters.includeUnpublished())   sql.append(" AND c.is_published = true");
```

Parametrlar doim bog'lanadi (`:param`) — SQL injection yo'q. `IN (:ids)`
uchun `List<Long>` beriladi, Spring o'zi kengaytiradi (PDO'da qo'lda `?,?,?`
yasashga to'g'ri kelganidan farqli).

Saralash (`sort` param): `popular` (student_count), `newest`, `price-asc`,
`price-desc` — switch bilan ORDER BY tanlanadi (parametr sifatida EMAS —
ORDER BY'ga parametr bog'lab bo'lmaydi, shuning uchun qiymatlar oldindan
validatsiyadan o'tgan).

### Kurs yaratish — tranzaksiya va slug

```java
@Transactional
void insert(CourseDto course) {
    // 1. Slug band bo'lsa -2, -3... qo'shib ketamiz
    for (int i = 2; ; i++) { ... }
    // 2. INSERT courses ... RETURNING id, created_at, version
    // 3. Har modul va dars INSERT (RETURNING id)
    // 4. Javob uchun aggregatlar hisoblanadi (DB'dan qayta o'qimasdan)
}
```

`@Transactional` — metod muvaffaqiyatli tugasa COMMIT, exception otilsa
ROLLBACK. Kurs 3 jadvalga yoziladi (courses, modules, lessons) — yarim
yozilgan holat qolmaydi.

### Tahrirlash: ReplaceModules strategiyasi

PATCH'da `modules` berilsa o'quv rejasi **butunlay almashtiriladi**:
eski modules/lessons DELETE (CASCADE), yangilari INSERT — yangi id'lar
bilan. Go'dagi qaror aynan shu (diff hisoblashdan ko'ra sodda); sotib
olganlarning kirishi enrollment tomonда qayta so'ralganda tiklanadi.

## 5.3. CourseService — dekoratsiya va facade

Kurs qatorida faqat `instructor_id` bor, frontend esa to'liq instruktor
obyektini kutadi. `decorate()` ro'yxatdagi barcha kurslar uchun:

```java
Set<Long> idSet = ...барcha instructorId'lar...;
List<UserSummary> users = userApi.findByIds(List.copyOf(idSet));  // BITTA batch chaqiruv

for (CourseDto c : list) {
    c.thumbnailColor = UiDefaults.thumbnailColor(c.id);
    c.instructor = new InstructorDto(...user.name() yoki "Instructor"...);
}
```

Go'da bu auth-service'ga HTTP chaqiruv edi; bu yerda `UserApi` facade.
Batch printsipi saqlangan (N+1 yo'q).

`CourseService` shuningdek `CourseApi`ni implement qiladi — boshqa
modullar uchun: `coursesByIds`, `coursesByInstructor`, `lessonsForCourse(s)`,
`lessonsByIds`, `incrementStudentCount`, `avgQuizScore`, `stats`,
`courseCountsByInstructor`. Bularning bari Go'dagi `/internal/*`
endpointlarning metodlarga aylangani.

## 5.4. CourseController — paywall

`GET /v1/courses/{idOrSlug}` — id ham slug ham qabul qilinadi (raqamga
o'girilsa id, bo'lmasa slug). Javobda modules + oxirgi 20 sharh.

Eng qiziq qism — **paywall** (`sanitizeCourseContent`): pullik darslarning
kontenti faqat kirish huquqi borlarga ko'rinadi:

```java
private void sanitizeCourseContent(CourseDto course) {
    if (course.modules == null || canModifyCourse(course)) return;  // egasi/admin hammasini ko'radi

    Set<Long> accessible = new HashSet<>();
    UserPrincipal claims = CurrentUser.get();
    if (claims != null) {
        try {
            accessible.addAll(enrollmentApi.accessibleLessonIds(claims.id(), course.id));
        } catch (RuntimeException e) {
            log.warn(...);   // FAIL CLOSED: xatoda faqat bepul darslar ochiq qoladi
        }
    }

    for (ModuleDto module : course.modules)
        for (LessonDto lesson : module.lessons)
            if (!lesson.isFree && !accessible.contains(lesson.id)) {
                lesson.content = "";
                lesson.contentUrl = "";
                lesson.locked = true;     // frontend qulf belgisini ko'rsatadi
            }
}
```

Kirish huquqi enrollment modulining `lesson_access` jadvalida —
`EnrollmentApi.accessibleLessonIds` facade orqali so'raladi (Go'da
`GET /internal/access`). "Fail closed" — huquqni tekshirib bo'lmasa kontent
YASHIRILADI (ochilmaydi).

Egalik tekshiruvi (`canModifyCourse`) — Laravel Policy'ning qo'lda versiyasi:

```java
claims.id() == course.instructorId || Roles.ADMIN.equals(claims.role())
```

PATCH/DELETE'da bu tekshiruvdan o'tmagan so'rov 403 oladi.

## 5.5. Qolgan controllerlar qisqacha

**CategoryController** — CRUD + 3 tilda nom (`name_uz/ru/en`), slug
`name_en`dan. Ota-bola daraxt (max 2 daraja — DB trigger tekshiradi).
`list()`da har kategoriya `courseCount` bilan chiqadi; ota kategoriya
soni = o'ziniki + bolalariniki (SQL'dan keyin Java'da yig'iladi).
Constraint xatolari validatsiya xabarlariga aylanadi
(duplicate slug → "a category with this name already exists" va h.k.).

**QuizController** — har kursga bitta quiz; frontend quizni **kurs id**
bilan so'raydi (`GET /v1/quizzes/{courseId}`). `PUT /v1/courses/{id}/quiz` —
quiz savollari bilan butunlay almashtiriladi (upsert `ON CONFLICT`).
Savol variantlari Postgres `text[]` massivida — bu JdbcClient bilan
ishlamaydi, shuning uchun `QuizRepository` bir joyda pastroq darajadagi
`JdbcTemplate` + `con.createArrayOf("text", ...)` ishlatadi.
Attempts: baholash clientda (correctIndex javobda ochiq — Go'dagi qaror),
server faqat score tarixini saqlaydi.

**ReviewController** — `POST /v1/courses/{id}/reviews`. Bitta user bitta
kursga bitta sharh: `ON CONFLICT (course_id, user_id) DO UPDATE`. User
ismi **snapshot** sifatida saqlanadi (`user_name` ustuni) — sharh
ko'rsatishda auth modulga murojaat kerak emas. Ism `UserApi.findByIds`dan
olinadi.

**InstructorController** — instruktor alohida jadval EMAS: "kamida bitta
published kursi bor user". Statistika `instructorStats()` SQL'idan
(kurslar soni, jami studentlar, o'rtacha reyting), ismlar `UserApi`dan.

Keyingi bo'lim: [enrollment modul](06-enrollment-modul.md).
