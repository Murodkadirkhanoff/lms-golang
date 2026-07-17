# 2. Spring Boot asoslari

## 2.1. Kirish nuqtasi

```java
// LmsApplication.java
@SpringBootApplication
@ConfigurationPropertiesScan
public class LmsApplication {
    public static void main(String[] args) {
        SpringApplication.run(LmsApplication.class, args);
    }
}
```

`main` — Java dasturining kirish nuqtasi (PHP'da `public/index.php`).
`SpringApplication.run(...)` quyidagilarni qiladi:

1. `uz.chashma.lms` paketi va uning ostidagi **hamma paketni skanerlaydi**;
2. annotationli klasslardan obyektlar ("bean"lar) yaratib containerga soladi;
3. ichki Tomcat web-serverni :4000 da ishga tushiradi;
4. Flyway migratsiyalarni o'tkazadi, DB pool ochadi.

`@SpringBootApplication` = component scan + auto-configuration.
Auto-configuration: classpath'da nima borligiga qarab o'zini sozlaydi —
`postgresql` driver bor → DataSource yaratadi; `spring-security` bor →
hamma endpointni himoyalaydi (biz `SecurityConfig`da o'zimizniki bilan
almashtiramiz).

## 2.2. DI container — Laravel service container'ning statik-tipli versiyasi

Laravel'da:

```php
// Laravel: constructor injection, container o'zi yechadi
class AuthController {
    public function __construct(private UserRepository $users) {}
}
```

Spring'da aynan shu g'oya:

```java
@RestController
class AuthController {
    private final UserRepository users;

    AuthController(UserRepository users) {   // container o'zi UserRepository beradi
        this.users = users;
    }
}
```

Farqlar:
- Laravel har so'rovda controllerni qayta yaratadi. Spring'da har bean
  **singleton** — ilova ishga tushganda BIR MARTA yaratiladi va hamma
  so'rovga xizmat qiladi. Shuning uchun bean maydonlarida so'rovga oid
  ma'lumot saqlanmaydi.
- Laravel'da binding'ni `AppServiceProvider`da yozish mumkin. Spring'da
  asosan annotation: klass ustiga `@Service`/`@Repository`/`@RestController`
  yozilsa — container avtomatik topib oladi.

Bean turlarini belgilaydigan annotationlar (hammasi bir xil ishlaydi, nomi
faqat semantika):

| Annotation | Ma'nosi | Loyihada misol |
|---|---|---|
| `@RestController` | HTTP controller, metodlari JSON qaytaradi | `AuthController`, `CourseController` |
| `@Repository` | DB qatlami | `UserRepository`, `CourseRepository` |
| `@Service` | Biznes-logika | `CourseService`, `EnrollmentService` |
| `@Configuration` + `@Bean` | Qo'lda bean yaratish (provider'dagi bind kabi) | `SecurityConfig` ichida `JwtService`, `PasswordEncoder` |

`@Bean` misoli — o'zimiz new qilib container'ga beramiz (Laravel'dagi
`$this->app->singleton(...)` ekvivalenti):

```java
// SecurityConfig.java
@Bean
public PasswordEncoder passwordEncoder() {
    return new BCryptPasswordEncoder(12);   // endi istalgan joyga inject qilinadi
}
```

Interfeys orqali injection — modullararo aloqaning kaliti:

```java
// AdminController (auth moduli) course modulining IMPLEMENTATSIYASINI bilmaydi:
AdminController(UserRepository users, CourseApi courseApi, EnrollmentApi enrollmentApi) { ... }
```

Container `CourseApi` interfeysini so'raganda uning yagona
implementatsiyasi `CourseService`ni topib beradi. Laravel'dagi
`$this->app->bind(CourseApi::class, CourseService::class)` avtomatik
bajarilgani kabi.

## 2.3. Routing — annotation orqali

Laravel'da `routes/api.php` alohida fayl. Spring'da route to'g'ridan-to'g'ri
metod ustida:

```java
@RestController
@RequestMapping("/v1")                       // prefiks (Route::prefix('v1'))
class AuthController {

    @PostMapping("/users")                   // POST /v1/users
    ResponseEntity<Map<String, Object>> register(@RequestBody RegisterRequest input) { ... }

    @PutMapping("/users/password")           // PUT /v1/users/password
    Map<String, Object> resetPassword(@RequestBody ResetPasswordRequest input) { ... }
}
```

Parametrlar:

| Annotation | Laravel ekvivalenti | Misol |
|---|---|---|
| `@PathVariable long id` | `Route::get('.../{id}')` + parametr | `show(@PathVariable long id)` |
| `@RequestParam String sort` | `$request->query('sort')` | kurs ro'yxati filtrlari |
| `@RequestBody RegisterRequest input` | `$request->json()` → DTO | JSON body avtomatik obyektga aylanadi |

Qaytarilgan obyekt (Map, record, DTO) avtomatik JSON'ga aylanadi
(Jackson — `json_encode`ning juda kuchli versiyasi). Status kod kerak
bo'lsa `ResponseEntity` bilan o'raladi:

```java
return ResponseEntity.status(HttpStatus.CREATED).body(Map.of("user", ..., "token", ...));
```

## 2.4. Record — DTO uchun qisqa sintaksis

Java 16+ da `record` — o'zgarmas (immutable) ma'lumot klassi:

```java
record LoginRequest(String email, String password) {}
```

Bu PHP 8 dagi constructor promotion + readonly'ga o'xshaydi:

```php
final class LoginRequest {
    public function __construct(
        public readonly ?string $email,
        public readonly ?string $password,
    ) {}
}
```

Maydonlarga `input.email()` deb murojaat qilinadi (getter avtomatik).
Loyihada barcha request DTO'lar — record. Muhim detal: JSON'da maydon
kelmasa record komponenti `null` bo'ladi, shuning uchun controller'larda
`orEmpty(...)` bilan `null → ""` normalizatsiya qilamiz (Go'da yo'q maydon
avtomatik `""` bo'lgani bilan tenglashish uchun).

### Snake_case requestlar

Frontend requestlarni snake_case yuboradi (`name_uz`, `course_id`).
Record komponentlari camelCase — moslashtirish uchun:

```java
@JsonNaming(PropertyNamingStrategies.SnakeCaseStrategy.class)
record CreateCategoryRequest(String nameUz, String nameRu, String nameEn, Long parentId) {}
```

DIQQAT: Spring Boot 4 **Jackson 3** ishlatadi — `@JsonNaming` importi
`tools.jackson.databind.annotation.JsonNaming` bo'lishi shart. Eski
`com.fasterxml.jackson.databind...` paketidan import qilinsa kompilyatsiya
bo'ladi-yu, **jimgina ishlamaydi** (classpath'da jjwt tufayli eski Jackson 2
ham bor). `@JsonIgnore`, `@JsonProperty`, `@JsonInclude` esa
`com.fasterxml.jackson.annotation` paketidan qoladi — Jackson 3 ularni
qo'llab-quvvatlaydi.

## 2.5. So'rovning umumiy yo'li

```
HTTP so'rov (Tomcat thread oladi)
  │
  ▼
Security filter chain                        ← Laravel middleware qatlami
  ├─ CORS filter
  ├─ JwtAuthFilter        — Bearer token bo'lsa parse qilib SecurityContext'ga qo'yadi
  └─ Authorization        — SecurityConfig'dagi qoidalar (/v1/admin/** → admin roli)
  │
  ▼
DispatcherServlet (router)                   ← route'ni topadi
  │
  ▼
Jackson: JSON body → record DTO              ← FormRequest'siz avtomatik
  │
  ▼
Controller metodi
  ├─ Validator (qo'lda, Go uslubida)         ← xato bo'lsa ValidationException
  ├─ Repository (JdbcClient, SQL)            ← PDO prepared statements kabi
  └─ Facade chaqiruvi (boshqa modul kerak bo'lsa)
  │
  ▼
Qaytgan obyekt → Jackson → JSON javob
  │
  ▼ (xato otilsa)
GlobalExceptionHandler (@RestControllerAdvice) ← app/Exceptions/Handler.php
  └─ {"error": ...} formatida javob
```

## 2.6. Konfiguratsiya — application.yml

Laravel'da `config/*.php` + `.env`. Spring'da bitta `application.yml`,
env o'zgaruvchilar `${NOMI:default}` sintaksisi bilan bog'lanadi:

```yaml
server:
  port: ${LMS_PORT:4000}          # env('LMS_PORT', 4000)

lms:                               # o'zimizning bo'lim
  env: ${LMS_ENV:development}
  jwt:
    secret: ${JWT_SECRET:}
    ttl: ${JWT_TTL:24h}            # avtomatik Duration'ga aylanadi
```

O'zimizning `lms.*` bo'limi tipli klassga bog'lanadi (config helper'siz):

```java
@ConfigurationProperties(prefix = "lms")
public record LmsProperties(String env, Jwt jwt, String corsTrustedOrigins) {
    public record Jwt(String secret, Duration ttl) {}
}
```

Endi istalgan bean `LmsProperties`ni constructor orqali oladi va
`props.jwt().ttl()` deb ishlatadi — `config('lms.jwt.ttl')`ning tip-xavfsiz
versiyasi.

Keyingi bo'lim: [shared modul](03-shared-modul.md) — validatsiya, xatolar
va security.
