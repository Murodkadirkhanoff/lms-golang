# LMS Backend — Java (Spring Boot 4)

Go backendning (`backend/`) to'liq Java porti — **modular monolith**: bitta
deploy qilinadigan ilova, ichida qat'iy chegaralangan `auth`, `course`,
`enrollment` modullari. API contract Go versiyasi bilan aynan bir xil —
mavjud Next.js frontend o'zgarishsiz ishlaydi.

**To'liq hujjatlar: [`docs/`](docs/README.md)** — PHP'dan kelgan backendchi
uchun 8 bo'limli qo'llanma (asoslar, Spring Boot, har modul alohida, DB
qatlami, so'rov hayoti misollari).

## Ishga tushirish

To'liq stack (Postgres + backend + frontend) repo root'dagi compose orqali:

```bash
# repo root'da; .env kerak (namuna: .env.example, JWT_SECRET majburiy)
docker compose up -d --build
```

Faqat backendni lokal (IDE/hot-reload) ishga tushirish — Postgres compose'dan
ishlab turgan bo'lsa yetarli:

```bash
cd backend-java/lms
JWT_SECRET=<secret> ./mvnw spring-boot:run     # Windows cmd: mvnw.cmd
```

Env o'zgaruvchilar (barchasi ixtiyoriy, default qavsda):
`LMS_PORT` (4000), `LMS_DB_URL` (jdbc:postgresql://localhost:5432/lms),
`LMS_DB_USER` (lms), `LMS_DB_PASSWORD` (devpassword), `JWT_SECRET`
(**majburiy**, ≥32 bayt — jjwt talabi), `JWT_TTL` (24h), `LMS_ENV`
(development), `CORS_TRUSTED_ORIGINS` (http://localhost:3000).

DB: `lms` bazasi, uch schema — `auth`, `course`, `enrollment` (Flyway
avtomatik yaratadi). Har modul faqat o'z schema'siga yozadi, schemalar
orasida FK yo'q — keyinchalik microservice'larga ajratishda har schema
alohida DB bo'lib chiqadi.

## Arxitektura

```
uz.chashma.lms
├── shared/          — texnik kod: Validator, xato envelope, JWT, security, UI defaults
├── auth/            — users, login/register, parol tiklash, admin panel
│   └── api/UserApi  — modul facade'i (boshqa modullar user'ni faqat shu orqali oladi)
├── course/          — kategoriyalar, kurslar, darslar, quizlar, sharhlar, instruktorlar
│   └── api/CourseApi, CourseDto...
└── enrollment/      — yozilishlar, buyurtmalar (mock to'lov), sertifikatlar, bildirishnomalar
    └── api/EnrollmentApi
```

Qoidalar:
- Modul boshqa modulning entity/repository'siga tegmaydi — faqat `api/`
  facade orqali (Go'dagi `/internal/*` REST endpointlarining o'rni).
- Ma'lumotlar qatlami Go'dagi kabi toza SQL (`JdbcClient`, ORM yo'q) —
  so'rovlar Go modellaridan 1:1 ko'chirilgan.
- Xato formati Go bilan bir xil: `{"error": "matn" | {field: message}}`;
  validatsiya xabarlari ham aynan.
- JWT formati Go bilan bir xil (HS256, issuer `lms.chashma.uz`) — Go
  yaratgan tokenlar Java'da ishlaydi va aksincha; bcrypt cost 12 bo'lgani
  uchun userlar ham almashinadi.

## Holat

Barcha Go endpointlari ko'chirilgan va smoke-testdan o'tgan: auth (register/
login/parol tiklash/admin), kurslar (CRUD, paywall, slug), kategoriyalar,
quizlar (upsert/attempts), sharhlar, instruktorlar, enroll/progress/sertifikat,
checkout (mock to'lov), buyurtmalar, bildirishnomalar, me/teaching statistika.

Keyingi bosqichlar:
1. Testlar (Testcontainers bilan integration testlar).
2. Spring Modulith bilan modul chegaralarini verify qilish.
3. Modullarni microservice'larga ajratish: facade → REST/gRPC, schema →
   alohida DB, Nginx gateway qaytadi (`backend/gateway/`).
