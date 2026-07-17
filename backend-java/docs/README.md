# LMS Java Backend — Hujjatlar

PHP (Laravel) dan kelgan backendchi uchun mo'ljallangan, ketma-ket o'qiladigan
qo'llanma. Har bo'limda Java/Spring tushunchalari PHP ekvivalentlari bilan
taqqoslanadi.

O'qish tartibi:

| # | Fayl | Mavzu |
|---|------|-------|
| 1 | [01-kirish-va-asoslar.md](01-kirish-va-asoslar.md) | Java vs PHP, Maven, loyiha tuzilishi, ishga tushirish |
| 2 | [02-spring-boot-asoslari.md](02-spring-boot-asoslari.md) | DI container, bean, annotationlar, so'rov hayoti, konfiguratsiya |
| 3 | [03-shared-modul.md](03-shared-modul.md) | Validator, xato formati, JWT, Security filter chain |
| 4 | [04-auth-modul.md](04-auth-modul.md) | Register/login, parol tiklash, admin panel, UserApi facade |
| 5 | [05-course-modul.md](05-course-modul.md) | Kategoriyalar, kurslar, paywall, quizlar, sharhlar |
| 6 | [06-enrollment-modul.md](06-enrollment-modul.md) | Enroll, progress, sertifikat, checkout, teaching stats |
| 7 | [07-database.md](07-database.md) | Flyway, schema-per-modul, JdbcClient, tranzaksiyalar |
| 8 | [08-sorov-hayoti-misollar.md](08-sorov-hayoti-misollar.md) | Ikki so'rovning boshdan-oxir yo'li (register, checkout) |

Umumiy kontekst: bu loyiha `backend/` dagi Go microservice backendning Java
porti — **modular monolith** ko'rinishida. API contract (URL'lar, JSON
shakllari, xato matnlari) Go versiyasi bilan aynan bir xil, shuning uchun
mavjud Next.js frontend o'zgarishsiz ishlaydi.
