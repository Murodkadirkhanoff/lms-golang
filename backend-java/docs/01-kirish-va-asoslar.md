# 1. Kirish va asoslar

## 1.1. Java PHP'dan nimasi bilan farq qiladi

| Tushuncha | PHP | Java |
|---|---|---|
| Ishlash modeli | Har so'rovga skript qayta ishga tushadi (FPM worker) | Ilova **bir marta** ishga tushadi va doim xotirada turadi; so'rovlar thread'larda xizmat qilinadi |
| Tiplash | Dinamik (type hint ixtiyoriy) | Qat'iy statik — har o'zgaruvchining tipi kompilyatsiyada tekshiriladi |
| Ishga tushirish | `php-fpm` + nginx | O'zi web-server (ichida Tomcat) — nginx shart emas |
| Kod yetkazish | Fayllarni serverga qo'yasan | `jar` fayl build qilinadi (`./mvnw package`) va `java -jar` bilan ishlaydi |
| Paket menejeri | Composer (`composer.json`) | Maven (`pom.xml`) |
| Namespace | `namespace App\Http\Controllers;` | `package uz.chashma.lms.auth;` — **papka tuzilishi paketga aynan mos bo'lishi shart** |
| Autoload | PSR-4 autoloader | Kompilyator + classpath (autoload tushunchasi yo'q — hammasi oldindan kompilyatsiya qilinadi) |

Eng muhim mental farq: PHP'da global holat so'rov tugashi bilan yo'qoladi.
Java'da ilova doim tirik — shuning uchun **holatni (state) ehtiyotkorlik
bilan saqlash** kerak: bitta obyekt minglab parallel so'rovlarga xizmat
qiladi. Shu sababli servislarimiz "stateless" (maydonlarida so'rovga oid
ma'lumot saqlamaydi).

## 1.2. Maven — Composer'ning ekvivalenti

`pom.xml` = `composer.json`. Farqi: XML formatda va build jarayonini ham
boshqaradi (kompilyatsiya, test, jar yig'ish).

`backend-java/lms/pom.xml` dan muhim qismlar:

```xml
<parent>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-parent</artifactId>
    <version>4.0.7</version>
</parent>
```

`parent` — Laravel'dagi `laravel/framework` ga o'xshaydi, lekin kuchliroq:
barcha kutubxonalarning **mos versiyalarini** markazlashtirib beradi. Shuning
uchun quyida dependency'larda versiya yozilmagan — parent belgilaydi:

```xml
<dependency>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-webmvc</artifactId>  <!-- HTTP/routing/JSON -->
</dependency>
<dependency>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-security</artifactId> <!-- auth middleware qatlami -->
</dependency>
<dependency>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-jdbc</artifactId>     <!-- PDO ekvivalenti -->
</dependency>
<dependency>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-flyway</artifactId>   <!-- migratsiyalar -->
</dependency>
```

"Starter" = tayyor paket to'plami. `starter-webmvc` o'zi bilan web-server
(Tomcat), routing, JSON (Jackson) olib keladi — Laravel'da bularning bari
framework'ka o'rnatilgan bo'lgani kabi.

`mvnw` / `mvnw.cmd` — **Maven wrapper**: Maven o'rnatilmagan bo'lsa ham
o'zi yuklab olib ishlatadi. `composer` global o'rnatilishiga ehtiyoj
qolmagani kabi.

Asosiy buyruqlar:

```bash
./mvnw compile           # kompilyatsiya (sintaksis + tip xatolarini ushlaydi)
./mvnw spring-boot:run   # dev-serverni ishga tushirish (php artisan serve)
./mvnw package           # target/lms-0.0.1-SNAPSHOT.jar yig'ish (deploy uchun)
```

## 1.3. Loyiha tuzilishi

```
backend-java/lms/
├── pom.xml                          # composer.json
├── mvnw, mvnw.cmd, .mvn/            # maven wrapper
└── src/main/
    ├── java/uz/chashma/lms/         # butun kod shu yerda (app/ kabi)
    │   ├── LmsApplication.java      # kirish nuqtasi (public/index.php)
    │   ├── shared/                  # umumiy texnik kod (app/Support kabi)
    │   │   └── security/            # JWT, filter, SecurityConfig
    │   ├── auth/                    # AUTH MODULI
    │   │   └── api/                 # modulning public interfeysi (facade)
    │   ├── course/                  # COURSE MODULI
    │   │   └── api/
    │   └── enrollment/              # ENROLLMENT MODULI
    │       └── api/
    └── resources/
        ├── application.yml          # config/*.php + .env birlashmasi
        └── db/migration/            # database/migrations
            ├── V1__auth_tables.sql
            ├── V2__course_tables.sql
            └── V3__enrollment_tables.sql
```

PHP'dagi `app/Http/Controllers`, `app/Models` kabi **tip bo'yicha** emas,
**modul bo'yicha** ajratilgan: `auth/` papkasida o'sha modulning controlleri,
repositoriyasi, DTO'si — hammasi birga. Bu "package by feature" deyiladi va
modular monolith uchun asos: modulni keyin alohida servisga ko'chirish oson.

### Ko'rinish (visibility) — modul chegarasining quroli

Java'da klass oldida `public` yozilmasa, u **package-private** bo'ladi —
faqat o'z paketi (papkasi) ichidan ko'rinadi:

```java
class UserRepository { ... }        // faqat uz.chashma.lms.auth ichida ko'rinadi
public interface UserApi { ... }    // hammaga ochiq (auth/api paketida)
```

Shu bilan qoidani kompilyator kuchi bilan ta'minlaymiz: **boshqa modul
auth'ning repositoriyasiga tegolmaydi** — faqat `UserApi` facade orqali
ishlaydi. PHP'da bunga o'xshash vosita yo'q (hammasi `public` yoki
konventsiyaga tayanasiz).

## 1.4. Modular monolith g'oyasi

Bu loyiha — Go'dagi 3 ta microservice'ning (auth :4001, course :4002,
enrollment :4003 + nginx :4000) **bitta ilovaga** birlashtirilgani:

| Go microservice dunyosi | Java monolith dunyosi |
|---|---|
| 3 alohida jarayon, 3 port | 1 jarayon, :4000 port |
| Servislar orasida HTTP (`/internal/*` + `X-Internal-Key`) | Oddiy metod chaqiruvi (facade interfeys) |
| 3 alohida DB (`lms_auth`, `lms_course`, `lms_enrollment`) | 1 DB (`lms`), 3 schema (`auth`, `course`, `enrollment`) |
| Nginx gateway routing qiladi | Kerak emas — hamma endpoint bitta ilovada |

Lekin microservice **intizomi** saqlangan:
- modullar bir-birining jadvallariga SQL yozmaydi (har modul o'z schema'si);
- schema'lar orasida FOREIGN KEY yo'q (`course.courses.instructor_id` —
  `auth.users`ga ishora, lekin FK emas);
- modullararo aloqa faqat `api/` paketidagi interfeyslar orqali.

Natija: keyinchalik modulni alohida microservice'ga chiqarish = facade
implementatsiyasini HTTP/gRPC client bilan almashtirish, schema'ni alohida
DB'ga ko'chirish. Kod logikasi o'zgarmaydi.

## 1.5. Ishga tushirish

```bash
# Postgres kerak (mavjud docker compose'dan lms-db):
cd backend && make up

# Ilova (gateway bilan port to'qnashmasligi uchun docker'dagi
# lms-gateway-full + Go servislarni to'xtat):
cd backend-java/lms
JWT_SECRET=dev-jwt-secret-not-for-production-8f2k1m ./mvnw spring-boot:run
```

Birinchi ishga tushishda Flyway o'zi schema'larni yaratib migratsiyalarni
o'tkazadi. Tekshirish: `curl localhost:4000/v1/healthcheck`.

Env o'zgaruvchilar (`.env` fayl yo'q — muhitdan o'qiladi, default'lar
`application.yml`da):

| Env | Default | Izoh |
|---|---|---|
| `LMS_PORT` | 4000 | HTTP port |
| `LMS_DB_URL` | jdbc:postgresql://localhost:5432/lms | JDBC ulanish satri |
| `LMS_DB_USER` / `LMS_DB_PASSWORD` | lms / devpassword | |
| `JWT_SECRET` | (bo'sh — **majburiy**) | Kamida 32 bayt (jjwt talabi) |
| `JWT_TTL` | 24h | Token amal qilish muddati |
| `LMS_ENV` | development | healthcheck'da ko'rinadi |
| `CORS_TRUSTED_ORIGINS` | http://localhost:3000 | Bo'shliq bilan ajratilgan ro'yxat |

Keyingi bo'lim: [Spring Boot asoslari](02-spring-boot-asoslari.md) —
DI container, annotationlar va so'rovning ilova ichidagi yo'li.
