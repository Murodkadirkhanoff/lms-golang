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

- [ ] **3. Rate limiting yo'q** — login/register endpointlariga cheklovsiz
  urinish mumkin (brute-force ochiq). `pkg/middleware`ga IP-based rate limiter
  qo'shish kerak.

- [ ] **4. Parolni tiklash real ishlamaydi** — reset token faqat server logiga
  yoziladi, SMTP ulanmagan (auth-service `tokens.go`da TODO). Email
  integratsiyasi kerak.

- [ ] **5. Token eskirganda UX buziladi** — JWT 24 soatlik, refresh yo'q;
  frontend axios interceptorida 401 da avtomatik logout/redirect yo'q.

- [ ] **6. Review'ga enrollment sharti yo'q** — istalgan login qilgan user
  sotib olmagan kursiga ham baho qo'yadi (takror UNIQUE bilan yopilgan, xolos).

## Funksional bo'shliqlar

- [ ] **7. Quiz builder UI yo'q** — backendda `PUT /courses/{id}/quiz` tayyor,
  studio'da quiz yaratish formasi yo'q. Hozir birorta kursga quiz qo'yib
  bo'lmaydi.

- [ ] **8. Review yozish UI yo'q** — `POST /courses/{id}/reviews` endpoint bor,
  frontend formasi yo'q.

- [ ] **9. Video/fayl yuklash infratuzilmasi yo'q** — dars videosi oddiy URL
  matn maydoni; kurs thumbnail bo'limi vizual, backendda fayl saqlash (S3/lokal)
  umuman yo'q.

- [ ] **10. Settings/Profile saqlanmaydi** — sahifalar soxta; profil yangilash
  endpointi backendda ham yo'q. Profile'dagi o'quv statistikasi hardcoded.

- [ ] **11. Katalog kategoriya filtri real rejimda ishlamaydi** — filtr
  hardcoded `CATEGORIES` ro'yxatidan "Development" kabi qiymat yuboradi, backend
  slug (`development`) kutadi → natija bo'sh. Yangi kategoriyalar filtrga
  chiqmaydi. Kurs sahifasida kategoriya yorlig'i `t("cat."+nom)` — dinamik
  kategoriyalarda xom i18n kaliti ko'rinadi. Filtrni `GET /categories`dan
  to'ldirish kerak.

## Kichikroq UX / texnik

- [ ] **12. Notifications** — bittalab "o'qildi" qilish yo'q (faqat read-all);
  navbar'da bell/badge yo'q.
- [ ] **13. Learn sahifasi tablari soxta** — notes saqlanmaydi, Q&A
  yuborilmaydi, resources ro'yxati hardcoded.
- [ ] **14. Sertifikat "Download" soxta** — PDF generatsiya yo'q.
- [ ] **15. Login redirect** — enroll/checkout login sahifasiga `?next=`siz
  yuboradi, login'dan keyin foydalanuvchi qaytib kelmaydi (RequireAuth buni
  qo'llaydi, shu joylar yo'q).
- [ ] **16. Pagination yo'q** — `/me/orders`, `/me/courses`, `/admin/users`
  unbounded.
- [ ] **17. Admin tayinlash yo'li yo'q** — userni admin qilish faqat DB orqali.
- [ ] **18. Checkout billing formasi** — "Amir Karimov" kabi hardcoded default
  qiymatlar qolgan.
