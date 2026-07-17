# Kamchiliklar ro'yxati (2026-07-15 auditi)

Muhimlik tartibida. Belgi: [ ] — ochiq, [x] — bajarilgan.

## Jiddiy (xavfsizlik / biznes-mantiq)

- [x] **1. Pullik kontent to'lovsiz ochiq** — `GET /courses/{id}` hamma darslarning
  `content` (matn) va `contentUrl` (video) maydonlarini sotib olmagan/login
  qilmagan foydalanuvchiga ham qaytaradi; learn sahifasida enrollment tekshiruvi
  yo'q. Yechim: backend javobida bepul bo'lmagan va foydalanuvchi kirish huquqi
  bo'lmagan darslarning kontentini olib tashlash (`locked: true`), learn
  sahifasida qulf UI.

- [x] **2. Checkout yaxlitlik tekshiruvlari yo'q** — bitta kursni qayta-qayta
  sotib olish mumkin (daromad ikki marta hisoblanadi), o'z kursini sotib olish
  mumkin. Frontend 8% soliq ko'rsatadi, backend soliqsiz saqlaydi — foydalanuvchi
  ko'rgan summa va buyurtma summasi mos emas.
  _Bajarildi: takroriy/egallangan kurs-darslar va o'z kursini xarid qilish
  backendda rad etiladi, savatdagi dublikatlar dedupe qilinadi, soliq qatori
  olib tashlandi (ko'rsatilgan = saqlangan), enrolled userga "davom etish"
  tugmasi ko'rsatiladi._

- [x] **3. Rate limiting yo'q** — _Bajarildi (2026-07-17): `RateLimitFilter` —
  IP bo'yicha token bucket (default 2 rps / burst 4, `RATE_LIMIT_*` env bilan
  sozlanadi) login/register/parol-tiklash endpointlarida; 429 + Go'dagi kabi
  `{"error": "rate limit exceeded"}`._

- [x] **4. Parolni tiklash real ishlamaydi** — _Bajarildi (2026-07-17):
  `spring-boot-starter-mail` + `MailService`; SMTP env orqali sozlanadi
  (`SMTP_HOST`...), compose'da Mailpit (web UI: http://localhost:8025), SMTP
  sozlanmagan bo'lsa havola logga yoziladi. Havola `FRONTEND_URL/reset-password`._

- [x] **5. Token eskirganda UX buziladi** — _Bajarildi (2026-07-17): axios
  interceptorida 401 kelsa token/user tozalanadi va `/login?next=<joriy sahifa>`
  ga yo'naltiriladi (login urinishining o'zi bundan mustasno)._

- [x] **6. Review'ga enrollment sharti yo'q** — _Bajarildi (2026-07-17):
  backend `EnrollmentApi.isEnrolled` tekshiradi (403), frontendda review formasi
  faqat enrolled userga ko'rinadi._

## Funksional bo'shliqlar

- [x] **7. Quiz builder UI yo'q** — _Bajarildi (2026-07-17): studio kurs
  tahrirlash sahifasida `QuizBuilder` (savollar/variantlar/to'g'ri javob,
  o'tish bali, vaqt) — `PUT /courses/{id}/quiz` ga ulandi._

- [x] **8. Review yozish UI yo'q** — _Bajarildi (2026-07-17): kurs sahifasida
  enrolled user uchun yulduzli baho + izoh formasi (`ReviewForm`)._

- [x] **9. Video/fayl yuklash infratuzilmasi yo'q** — _Bajarildi (2026-07-17):
  `POST /v1/uploads` (multipart, auth, video/rasm oq ro'yxati, 512MB limit),
  fayllar docker volume'da (`/app/uploads`), `/uploads/**` statik beriladi.
  Kursga `thumbnail_url` ustuni (V4), studio formada rasm yuklash + dars
  videosi real yuklanadi (progress bilan). S3'ga o'tish `UploadController`ni
  almashtirish bilan bo'ladi._

- [x] **10. Settings/Profile saqlanmaydi** — _Bajarildi (2026-07-17): backend
  `GET /v1/me`, `PUT /v1/me/profile` (ism), `PUT /v1/me/password` (joriy parol
  tasdig'i bilan); profile/settings sahifalari ulandi, o'quv statistikasi
  `GET /me/stats` dan (real)._

- [x] **11. Katalog kategoriya filtri real rejimda ishlamaydi** — _Bajarildi
  (2026-07-17): filtr `GET /categories` dan quriladi (qiymat — slug), kategoriya
  yorlig'i barcha joylarda dinamik nomdan (`useCategoryName`, 3 til)._

## Kichikroq UX / texnik

- [x] **12. Notifications** — _Bajarildi (2026-07-17): `POST
  /me/notifications/{id}/read` (bittalab), navbar'da bell + o'qilmaganlar soni
  badge'i, sahifadagi belgilashlar endi serverga yoziladi._
- [x] **13. Learn sahifasi tablari soxta** — _Bajarildi (2026-07-17): notes
  localStorage'da dars kesimida saqlanadi; resources real dars kontentidan;
  Q&A backendda saqlanadi (`course.lesson_questions` V5, GET/POST
  `/lessons/{id}/questions`). Instruktor javob threadi — kelajak ishi._
- [x] **14. Sertifikat "Download" soxta** — _Bajarildi (2026-07-17): `GET
  /me/certificates/{id}/download` — OpenPDF bilan A4 landscape PDF (ism, kurs,
  sana, ID); frontend tugmasi blob orqali yuklab oladi._
- [x] **15. Login redirect** — _Bajarildi (2026-07-17): enroll/checkout endi
  `?next=` bilan yuboradi; login sahifasi allaqachon qaytaradi edi._
- [x] **16. Pagination** — _Bajarildi (2026-07-17): `/me/orders` va
  `/me/courses` backendda sahifalandi (page/pageSize/total envelope);
  purchases, my-courses va admin/users sahifalarida Pagination UI._
- [x] **17. Admin tayinlash yo'li yo'q** — _Bajarildi (2026-07-17): `PATCH
  /v1/admin/users/{id}/role` (o'z rolini o'zgartirish taqiqlangan); admin users
  jadvalida rol select'i._
- [x] **18. Checkout billing formasi** — _Bajarildi (2026-07-17): ism/email
  login qilgan userdan prefill, hardcoded qiymatlar olib tashlandi._

## Keyingi g'oyalar (yangi)

- [ ] Q&A'da instruktor javoblari (threading) va Q&A'ga enrollment sharti.
- [ ] Yuklangan videolarni paywall ortiga olish (hozir URL topilsa ochiq —
  UUID'li nom taxmin qilib bo'lmaydi, lekin signed URL to'g'riroq).
- [ ] Notification prefs / 2FA / billing kartalari (settings'dagi qolgan
  bo'limlar vizual).
- [ ] Rate limiter'ni k8s'da ko'p replica uchun markazlashtirish (Redis).
