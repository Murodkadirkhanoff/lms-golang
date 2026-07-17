# 4. `auth` modul — foydalanuvchilar va autentifikatsiya

```
auth/
├── api/
│   └── UserApi.java          # PUBLIC facade — boshqa modullar uchun yagona eshik
├── User.java                 # auth.users qatori (package-private!)
├── UserResponse.java         # javob DTO (id, createdAt, name, email, role)
├── UserRepository.java       # SQL qatlami
├── PasswordResetTokenRepository.java
├── UserApiImpl.java          # facade implementatsiyasi
├── AuthController.java       # register/login/parol tiklash
└── AdminController.java      # /v1/admin/* (users ro'yxati, statistika)
```

DB: `auth` schema — `users`, `password_reset_tokens` jadvallari.

## 4.1. User — entity emas, oddiy qator-klass

```java
class User {          // package-private: faqat auth moduli ko'radi
    long id;
    Instant createdAt;
    String name;
    String email;
    byte[] passwordHash;   // DB'da bytea — Go bilan mos (userlar almashinadi)
    String role;
    int version;           // optimistic locking uchun
}
```

Eloquent'dagi Model emas — hech qanday sehr yo'q, shunchaki ma'lumot
tashuvchi (Go struct kabi). SQL bilan `UserRepository` ishlaydi.
JSON'ga esa `User` emas, `UserResponse` record chiqadi — parol hash
tasodifan ham javobga tushmasligi kafolatlanadi (Laravel `$hidden`
o'rniga umuman alohida DTO).

## 4.2. UserRepository — SQL qatlami

`JdbcClient` — Spring'ning zamonaviy SQL clienti (PDO prepared
statement'larining qulay o'rami). Misol:

```java
Optional<User> findByEmail(String email) {
    return jdbc.sql("""
            SELECT id, created_at, name, email, password_hash, role, version
            FROM auth.users
            WHERE email = :email AND deleted_at IS NULL
            """)
            .param("email", email)
            .query(FULL_MAPPER)     // ResultSet -> User (qo'lda mapping)
            .optional();            // 0 yoki 1 qator
}
```

E'tibor beriladigan joylar:

- **`deleted_at IS NULL`** — soft delete har so'rovda qo'lda (Eloquent
  SoftDeletes trait'i avtomatik qo'shganidan farqli, bu yerda oshkora).
- **`email` ustuni `citext`** — Postgres'ning case-insensitive text tipi,
  `WHERE email = ...` katta-kichik harfni farqlamaydi. `strtolower()`
  kerak emas.
- **INSERT ... RETURNING** — Postgres'ga xos: yozish bilan birga id,
  created_at, version qaytadi (alohida `lastInsertId()` yo'q).
- **Optimistic locking** (`updatePassword`):

```java
UPDATE auth.users SET password_hash = :hash, version = version + 1
WHERE id = :id AND version = :version
```

0 qator yangilansa — boshqa so'rov oldin o'zgartirgan → `EditConflictException`
(409). Bu SELECT ... FOR UPDATE qulfisiz parallel yozishdan himoya.

- **Sahifalash** (`list`): `count(*) OVER()` window funksiyasi umumiy sonni
  har qatorda qaytaradi — alohida `SELECT COUNT(*)` so'rovi kerak emas.

Duplicate email `PgErrors` orqali ushlanadi (3.5-bo'limga qarang) va
`DuplicateEmailException`ga aylanadi.

## 4.3. AuthController — to'rt oqim

### POST /v1/users — register

```
1. null'larni "" ga normalizatsiya (orEmpty)
2. Validatsiya: name (<=500 bayt), email (regex), parol (8..72 bayt)
3. User yaratish: role doim "student", bcrypt(12) hash
4. INSERT; duplicate email -> 422
5. JWT yaratish va javob: 201 {"user": {...}, "token": "..."}
```

Bcrypt: `passwordEncoder.encode(password)` `$2a$12$...` satr qaytaradi,
u `bytes` ko'rinishida bytea ustunga yoziladi — Go
(`golang.org/x/crypto/bcrypt`, cost 12) bilan to'liq mos: bir DB'dagi
userlar ikkala backend'da ham login qila oladi.

### POST /v1/tokens/authentication — login

Email bo'yicha topish → `passwordEncoder.matches(raw, hash)` → JWT.
Muhim xavfsizlik detali: user topilmasa ham, parol xato bo'lsa ham
javob **bir xil**: 401 `"invalid authentication credentials"` —
email bazada borligini oshkor qilmaymiz.

### POST /v1/tokens/password-reset — unutilgan parol

```java
Optional<User> user = users.findByEmail(email);
if (user.isPresent()) {
    String token = tokens.create(user.get().id, Duration.ofMinutes(45));
    log.info("password reset token created ... token={}", token);  // SMTP hali yo'q
}
return Map.of("message", "if the email address exists, a password reset link will be sent");
```

Javob **doim** 200 va bir xil matn — email mavjudligini bilib bo'lmaydi
(user enumeration himoyasi). Token hozircha log'ga yoziladi (Go'da ham
shunday — SMTP keyin ulanadi).

### PUT /v1/users/password — yangi parol o'rnatish

Token tekshiriladi → parol yangilanadi → userning **barcha** reset
tokenlari o'chiriladi. Metod `@Transactional` — uch yozuv amali bitta
tranzaksiyada (Laravel `DB::transaction(fn() => ...)` ekvivalenti,
faqat annotation ko'rinishida).

## 4.4. PasswordResetTokenRepository — xavfsiz token saqlash

Laravel `password_reset_tokens` jadvaliga o'xshash g'oya, lekin qat'iyroq:

```
1. 16 tasodifiy bayt (SecureRandom)
2. plaintext = Base32(bytes)          <- foydalanuvchiga beriladigan token
3. DB'ga faqat SHA-256(plaintext) hash yoziladi
```

DB o'g'irlansa ham tokenlarni ishlatib bo'lmaydi (hashdan plaintext
tiklanmaydi). Tekshirishda kelgan tokenning SHA-256'i DB'dagi hash bilan
solishtiriladi + `expiry > NOW()`.

Base32 encoder qo'lda yozilgan (~15 qator) — Java stdlib'da Base32 yo'q,
Go `base32.StdEncoding` (paddingsiz) bilan bir xil natija berishi kerak
edi (alfavit `A-Z2-7`).

## 4.5. AdminController — modullararo aggregatsiya

`GET /v1/admin/users` — sahifalangan ro'yxat, lekin har userga boshqa
modullardan hisoblagichlar qo'shiladi:

```java
Map<Long, Integer> createdCounts  = courseApi.courseCountsByInstructor(ids);   // course modulidan
Map<Long, Integer> enrolledCounts = enrollmentApi.enrollmentCountsByUser(ids); // enrollment modulidan
```

Go'da bular ikki HTTP chaqiruv edi (`/internal/course-counts`,
`/internal/enrollment-counts`) — monolithda oddiy metod chaqiruvi, lekin
**batch** printsipi saqlangan: N ta user uchun N ta so'rov emas, bitta
`ids` ro'yxati bilan bitta chaqiruv (N+1 muammosining oldi olingan).

Javob elementi frontend `AdminUser` shakliga aynan mos:

```java
record AdminUser(long id, String name, String email, String role, String avatarColor,
                 String joinedAt,   // "yyyy-MM-dd"
                 String status,     // doim "active"
                 int coursesCreated, int coursesEnrolled) {}
```

`GET /v1/admin/stats` — `totalUsers` (lokal) + `totalCourses`/
`activeInstructors` (CourseApi) + `revenue` (EnrollmentApi).

Route himoyasi controller'da EMAS — `SecurityConfig`da
(`/v1/admin/** → hasRole(admin)`). Controller chaqirilganidayoq user admin
ekani kafolatlangan.

## 4.6. UserApi — modul facade'i

```java
public interface UserApi {
    List<UserSummary> findByIds(List<Long> ids);
    record UserSummary(long id, Instant createdAt, String name, String email, String role) {}
}
```

Boshqa modullar (course sharh/instruktor ismlari uchun) user ma'lumotini
FAQAT shu orqali oladi. `UserSummary`da **parol hash yo'q** — facade
tashqariga minimal ma'lumot chiqaradi. Go'da bu `GET /internal/users?ids=1,2`
endpointi edi; microservice'ga ajratilganda bu interfeys implementatsiyasi
yana HTTP clientga aylanadi, chaqiruvchi kod esa o'zgarmaydi.

Keyingi bo'lim: [course modul](05-course-modul.md).
