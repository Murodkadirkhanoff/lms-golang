# 7. Database qatlami

## 7.1. Nega ORM (Hibernate/JPA) ishlatilmagan?

Spring dunyosida standart — Hibernate (Eloquent'ning "katta akasi").
Bu loyihada esa **toza SQL** (`JdbcClient`) tanlangan, sabablari:

1. Go backend toza SQL'da yozilgan — so'rovlarni 1:1 ko'chirish xatti-
   harakat aynan bir xilligini kafolatlaydi (aggregatlar, LATERAL join,
   `count(*) OVER()`, `ON CONFLICT` — bularni ORM'da qayta ifodalash
   xato manbai);
2. ORM'siz nima DB'ga ketayotgani doim ko'rinib turadi;
3. PHP'dan kelganlar uchun PDO'ga tanish model.

Taqqoslash:

| PDO | JdbcClient |
|---|---|
| `$pdo->prepare("... WHERE id = ?")` | `jdbc.sql("... WHERE id = :id")` |
| `$stmt->execute([$id])` | `.param("id", id)` |
| `$stmt->fetch()` | `.query(MAPPER).optional()` |
| `$stmt->fetchAll()` | `.query(MAPPER).list()` |
| `$stmt->rowCount()` | `.update()` (yangilangan qatorlar soni) |

## 7.2. RowMapper — fetch natijasini obyektga aylantirish

PDO `FETCH_CLASS`ga o'xshaydi, lekin qo'lda va tip-xavfsiz:

```java
private static final RowMapper<User> FULL_MAPPER = (rs, rowNum) -> {
    User user = new User();
    user.id = rs.getLong("id");
    user.createdAt = rs.getObject("created_at", OffsetDateTime.class).toInstant();
    user.email = rs.getString("email");
    user.passwordHash = rs.getBytes("password_hash");   // bytea -> byte[]
    ...
    return user;
};
```

Muhim detal — vaqt bilan ishlash: `timestamptz` ustuni
`rs.getObject(..., OffsetDateTime.class)` bilan o'qiladi va `Instant`ga
o'giriladi. `Instant` JSON'da `"2026-07-16T10:35:02Z"` (ISO-8601, UTC)
ko'rinishida chiqadi — Go `time.Time` marshal formati bilan bir xil.

## 7.3. Flyway — migratsiyalar

Laravel migratsiyalari PHP klasslar; Flyway'da — **oddiy SQL fayllar**,
nomlash qat'iy:

```
src/main/resources/db/migration/
├── V1__auth_tables.sql          # V<versiya>__<tavsif>.sql (ikki underscore!)
├── V2__course_tables.sql
└── V3__enrollment_tables.sql
```

Ilova ishga tushganda Flyway `flyway_schema_history` jadvalini tekshiradi
va yangi fayllarni tartib bilan o'tkazadi (`php artisan migrate` avtomatik
bajarilgani kabi). Rollback fayllari yo'q — faqat oldinga (kerak bo'lsa
yangi migratsiya bilan tuzatiladi).

`application.yml`dagi sozlama:

```yaml
spring:
  flyway:
    schemas: auth, course, enrollment   # shu schema'larni yaratadi ham
    create-schemas: true
```

## 7.4. Schema-per-modul

Bitta `lms` bazasi, uch schema. Har jadval nomi migratsiyada va so'rovlarda
to'liq yoziladi: `auth.users`, `course.courses`, `enrollment.orders`.

Qat'iy qoidalar (microservice'ga tayyorgarlik):

1. **Modul faqat o'z schema'siga SQL yozadi.** `CourseRepository`da
   `enrollment.` so'zi uchramaydi va aksincha.
2. **Schema'lar orasida FK yo'q.** `course.courses.instructor_id` —
   shunchaki bigint. Yaxlitlikni kod ta'minlaydi (Go'da servislararo FK
   bo'lishi mumkin emas edi — o'sha model saqlangan).
3. **Ma'lumot kerak bo'lsa — facade.** JOIN o'rniga
   `userApi.findByIds(...)` va Java'da birlashtirish.

Bunun evaziga: modulni ajratishda `pg_dump --schema=course` bilan
ma'lumot ko'chiriladi, kod esa deyarli o'zgarmaydi.

## 7.5. Tranzaksiyalar

`@Transactional` annotation — metod atrofida BEGIN/COMMIT/ROLLBACK:

```java
@Transactional
void insert(CourseDto course) {   // kurs + modullar + darslar — hammasi yoki hech
    ...
}
```

Laravel'dagi `DB::transaction(function () { ... })`ning deklarativ
ko'rinishi. Exception otilsa avtomatik ROLLBACK. Loyihada qo'llangan
joylar: kurs yaratish/modullarni almashtirish, quiz upsert, order+items
yozish, parol reset (yangilash + tokenlarni o'chirish).

Bilish kerak bo'lgan nozik joy: `@Transactional` **proxy** orqali ishlaydi —
xuddi shu klass ichidan `this.metod()` deb chaqirilsa tranzaksiya
ochilmaydi. Shuning uchun u faqat tashqaridan chaqiriladigan metodlarga
qo'yilgan.

## 7.6. Postgres'ning ishlatilgan xususiyatlari (lug'at)

| Xususiyat | Qayerda | Nima uchun |
|---|---|---|
| `GENERATED ALWAYS AS IDENTITY` | hamma id | AUTO_INCREMENT'ning SQL-standart shakli |
| `citext` | users.email | case-insensitive unique email |
| `bytea` | password_hash, token hash | binar ma'lumot |
| `text[]` | quiz_questions.options | massiv ustun (JSON o'rniga) |
| `INSERT ... RETURNING` | hamma insert | id/created_at'ni alohida so'rovsiz olish |
| `ON CONFLICT DO NOTHING/UPDATE` | enroll, review, sertifikat, quiz | idempotent yozish (race'siz upsert) |
| `count(*) OVER()` | sahifalash | umumiy son + qatorlar bitta so'rovda |
| `LEFT JOIN LATERAL` | kurs aggregatlari | har qator uchun subquery |
| trigger | category depth, order total | DB darajasidagi invariantlar |
| partial index | `WHERE deleted_at IS NULL` | faqat tirik qatorlar indeksi |
| `to_char(date_trunc(...))` | oylik daromad | oy bo'yicha guruhlash |

## 7.7. Soft delete va optimistic locking

**Soft delete** (users, courses, categories): `deleted_at` ustuni,
o'chirish = `SET deleted_at = NOW()`, har SELECT'da `AND deleted_at IS NULL`.
Kurs qattiq o'chirilmaydi chunki enrollment modulida unga ishoralar bor.
(Kategoriya esa qattiq o'chiriladi — Go'dagi xatti-harakat.)

**Optimistic locking** (users, courses, categories, quizzes): `version`
ustuni. UPDATE doim `WHERE ... AND version = :version` + `version = version + 1`.
0 qator o'zgarsa → `EditConflictException` → 409 "please try again".
Ikki admin bir vaqtda bitta kursni tahrirlasa, ikkinchisining yozuvi
birinchisinikini indamay ustidan yozib yubormaydi.

Keyingi bo'lim: [so'rov hayoti — misollar](08-sorov-hayoti-misollar.md).
