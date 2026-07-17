# 3. `shared` modul — umumiy texnik kod

`uz.chashma.lms.shared` — Laravel'dagi `app/Support` + `app/Exceptions` +
`app/Http/Middleware` aralashmasi. Bu yerda **domen logikasi yo'q** — faqat
texnik infratuzilma. Qoida: auth/course/enrollment shu paketga bog'lanishi
mumkin, teskarisi mumkin emas.

Fayllar ro'yxati:

```
shared/
├── LmsProperties.java          # tipli konfiguratsiya (2-bo'limda ko'rildi)
├── Validator.java              # qo'lda validatsiya (Go uslubida)
├── ValidationException.java    # 422 xatosi
├── BadRequestException.java    # 400
├── NotFoundException.java      # 404
├── InvalidCredentialsException.java  # 401 (login xato)
├── NotPermittedException.java  # 403
├── EditConflictException.java  # 409 (optimistic lock)
├── GlobalExceptionHandler.java # markaziy xato handler
├── JsonErrorWriter.java        # filter ichidan JSON xato yozish
├── PgErrors.java               # Postgres constraint nomini ajratish
├── UiDefaults.java             # avatarColor/thumbnailColor palitrasi
├── Ids.java                    # "1,2,3" -> List<Long>
├── HealthController.java       # GET /v1/healthcheck
└── security/
    ├── JwtService.java         # token yaratish/tekshirish
    ├── JwtClaims.java          # (userId, role) record
    ├── JwtAuthFilter.java      # Bearer tokenni o'qiydigan middleware
    ├── SecurityConfig.java     # himoya qoidalari (routes/middleware config)
    ├── UserPrincipal.java      # joriy user (id, role)
    ├── CurrentUser.java        # auth()->user() ekvivalenti
    ├── Roles.java              # "student"/"instructor"/"admin" konstantalar
    └── InvalidTokenException.java
```

## 3.1. Validator — nega Laravel'dagidek emas?

Laravel'da `$request->validate([...])` deklarativ. Bu yerda esa Go
backend'dagi `pkg/validator`ning aynan porti — chunki **xato xabarlari
va formati frontend bilan contract**: `{"error": {"maydon": "xabar"}}`,
matnlar aynan bir xil bo'lishi shart ("must be provided",
"must be at least 8 bytes long"...).

```java
Validator v = new Validator();
v.check(!email.isEmpty(), "email", "must be provided");
v.check(Validator.matches(email, Validator.EMAIL_RX), "email", "must be a valid email address");
v.throwIfInvalid();    // xato bo'lsa ValidationException -> 422
```

Ichida oddiy `LinkedHashMap<String, String>` — birinchi xato saqlanadi
(`putIfAbsent`), tartib buzilmaydi. `throwIfInvalid()` chaqirilganda xatolar
bo'lsa `ValidationException` otiladi, uni `GlobalExceptionHandler` 422 ga
aylantiradi.

Yordamchi statik metodlar:
- `byteLength(s)` — Go `len()` baytlarda o'lchagani uchun chegara
  tekshiruvlari ham baytda (UTF-8 da kirill harf 2 bayt!);
- `permitted(value, "uz", "ru", "en")` — Laravel `in:` qoidasi;
- `orEmpty(s)` — `null → ""` (JSON'da kelmagan maydonni Go semantikasiga
  tenglashtirish).

## 3.2. Xato ierarxiyasi va GlobalExceptionHandler

Laravel'da `app/Exceptions/Handler.php` da `render()` yozasiz. Spring'da
buning nomi `@RestControllerAdvice`:

```java
@RestControllerAdvice
public class GlobalExceptionHandler {

    @ExceptionHandler(ValidationException.class)
    public ResponseEntity<Map<String, Object>> failedValidation(ValidationException ex) {
        return error(HttpStatus.UNPROCESSABLE_ENTITY, ex.errors());   // 422 {"error": {field: msg}}
    }

    @ExceptionHandler(Exception.class)   // eng oxirgi to'r — kutilmagan xatolar
    public ResponseEntity<Map<String, Object>> serverError(Exception ex) {
        log.error("unhandled exception", ex);
        return error(HttpStatus.INTERNAL_SERVER_ERROR,
                "the server encountered a problem and could not process your request");
    }
}
```

Istalgan controller/repositorydan otilgan exception shu yerga keladi.
Qamrab olingan holatlar:

| Exception | HTTP | Qachon |
|---|---|---|
| `ValidationException` | 422 | Validator xatolari, duplicate email/slug |
| `BadRequestException` | 400 | Yaroqsiz query param |
| `HttpMessageNotReadableException` | 400 | Buzuq JSON body (Jackson otadi) |
| `NotFoundException`, `NoResourceFoundException`, `MethodArgumentTypeMismatchException` | 404 | Topilmadi / `{id}` raqam emas (Go ham 404 qaytaradi) |
| `InvalidCredentialsException` | 401 | Login parol xato |
| `NotPermittedException` | 403 | Birovning kursini o'zgartirish |
| `EditConflictException` | 409 | Optimistic lock (version mos kelmadi) |
| `Exception` | 500 | Qolgan hammasi (matn Go bilan bir xil) |

Bu yondashuvning qulayligi: controller'lar "happy path"ni yozadi, xato
holatlar exception bilan yuqoriga otiladi — Go'dagi har qadamda
`if err != nil` yozishdan farqli.

## 3.3. JWT — JwtService

Go `pkg/auth/jwt.go`ning porti, **format bir xil**: HS256, issuer
`lms.chashma.uz`, `sub` = user id (string), `role` claim. Shu tufayli Go
yaratgan token Java'da ishlaydi va aksincha (secret bir xil bo'lsa).

```java
public String newToken(long userId, String role) {
    Instant now = Instant.now();
    return Jwts.builder()
            .subject(Long.toString(userId))
            .issuer(ISSUER)
            .issuedAt(Date.from(now))
            .expiration(Date.from(now.plus(ttl)))
            .claim("role", role)
            .signWith(key, Jwts.SIG.HS256)
            .compact();
}
```

Kutubxona — `jjwt` (PHP'dagi `firebase/php-jwt` ekvivalenti). Bitta nozik
joy: jjwt HS256 uchun **kamida 32 baytlik** secret talab qiladi, aks holda
ishga tushishda xato beradi. `parse()` yaroqsiz/eskirgan tokenda
`InvalidTokenException` otadi — filter uni 401 ga aylantiradi.

## 3.4. Security — Laravel middleware'ning Spring versiyasi

### JwtAuthFilter

`OncePerRequestFilter` — har so'rovda bir marta ishlaydigan middleware:

```java
protected void doFilterInternal(request, response, chain) {
    String header = request.getHeader("Authorization");
    if (header == null) { chain.doFilter(...); return; }   // anonim — davom

    // "Bearer <token>" bo'lmasa yoki token buzuq bo'lsa — DARHOL 401:
    // {"error":"invalid or missing authentication token"} + WWW-Authenticate: Bearer

    JwtClaims claims = jwtService.parse(token);
    // Muvaffaqiyat: SecurityContext'ga joylaymiz (auth()->setUser() kabi)
    SecurityContextHolder.getContext().setAuthentication(
        new UsernamePasswordAuthenticationToken(
            new UserPrincipal(claims.userId(), claims.role()),
            null,
            List.of(new SimpleGrantedAuthority("ROLE_" + claims.role()))));
    chain.doFilter(request, response);
}
```

Muhim semantika (Go bilan bir xil): header **yo'q** bo'lsa so'rov anonim
davom etadi (ochiq endpointlar ishlashi uchun); header **bor lekin buzuq**
bo'lsa — darhol 401.

### SecurityConfig — himoya qoidalari

Laravel'da `Route::middleware('auth')->group(...)` ga to'g'ri keladi:

```java
.authorizeHttpRequests(a -> a
    .requestMatchers("/v1/admin/**").hasRole(Roles.ADMIN)          // faqat admin
    .requestMatchers(HttpMethod.POST, "/v1/courses").authenticated()
    .requestMatchers(HttpMethod.PATCH, "/v1/courses/*").authenticated()
    .requestMatchers(HttpMethod.DELETE, "/v1/courses/*").authenticated()
    .requestMatchers("/v1/courses/*/reviews", "/v1/courses/*/quiz",
            "/v1/quizzes/*/attempts", "/v1/courses/*/enroll",
            "/v1/enrollments/**", "/v1/me/**").authenticated()
    .anyRequest().permitAll());                                     // qolgani ochiq
```

Yana: CSRF o'chirilgan (SPA + token auth, sessiya yo'q), sessiya STATELESS,
CORS `CORS_TRUSTED_ORIGINS`dan. Qoidalar buzilganda javoblar Go bilan aynan:

- anonim himoyalangan joyga kirsa → 401 `"you must be authenticated to access this resource"`;
- roli yetmasa → 403 `"your user account doesn't have the necessary permissions..."`.

Bu javoblar `exceptionHandling(...)` dagi ikki lambda'da yoziladi. Ular
MVC'dan tashqarida (filter darajasida) ishlagani uchun JSON'ni
`JsonErrorWriter` qo'lda yozadi.

### CurrentUser — auth()->user()

```java
UserPrincipal claims = CurrentUser.get();   // null bo'lishi mumkin (anonim)
claims.id();      // user id (JWT sub'dan)
claims.role();    // "student" | "instructor" | "admin"
```

`SecurityContextHolder` — thread-local saqlagich: har so'rov o'z thread'ida
o'z userini ko'radi. DIQQAT: bu **DB'dagi user emas**, faqat JWT'dagi
ma'lumot — DB'ga murojaatsiz (Go ham xuddi shunday qiladi).

## 3.5. Mayda yordamchilar

**UiDefaults** — DB'da saqlanmaydigan, frontend kutuvchi UI ranglar:

```java
UiDefaults.avatarColor(userId)      // palette[id % 6]      -> "bg-indigo-200"...
UiDefaults.thumbnailColor(courseId) // palette[(id+3) % 6]
```

Deterministik: bir xil id doim bir xil rang. Go `pkg/uidefaults` bilan
palitra va formulalar aynan bir xil (aks holda UI ranglari "sakrab" qolardi).

**Ids** — `?ids=1,2,3` query paramni `List<Long>`ga aylantiradi; yaroqsiz
bo'laklarni indamay tashlaydi (Go `ReadIDList` kabi).

**PgErrors** — DB constraint xatolarini aniqlash. PHP'da
`PDOException::getCode() == 23000` deb tekshirganingizga o'xshaydi, lekin
biz aniq **constraint nomini** olamiz:

```java
try {
    users.insert(user);
} catch (DataAccessException ex) {
    if ("users_email_key".equals(PgErrors.constraint(ex))) {
        throw new DuplicateEmailException();       // -> 422 {"email": "...already exists"}
    }
    throw ex;
}
```

Bu Go'dagi `pq.Error.Constraint` tekshiruvining porti — unique/FK
buzilishlarini INSERT'dan OLDIN SELECT qilib tekshirmasdan, atomik
ravishda ushlaymiz.

Keyingi bo'lim: [auth modul](04-auth-modul.md).
