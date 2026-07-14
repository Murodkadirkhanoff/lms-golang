# LMS — Backend API Endpoints

Derived from the frontend `src/services/*` calls and the UI actions they back.
Endpoints are grouped into **bounded contexts** so each group can become its own
microservice later. Until then, run them all as one Chi mux under `/v1`.

## Conventions

- **Base prefix:** `/v1`
- **Auth:** JWT `Authorization: Bearer <token>` (issued by Identity). `🔒` = requires auth, `👤` = must be resource owner, `⚙️` = admin, `🎓` = instructor/owner of the course.
- **Error envelope:** `{ "error": <string | object> }` (already parsed by `src/lib/axios.ts`).
- **List envelope:** `{ "items": [...], "page", "pageSize", "total" }` (matches `Paginated<T>`).
- **IDs:** courses use `slug` for public reads, numeric `id` for owner/admin writes.
- **`/me/*`** = current authenticated user (server derives id from JWT, never trust a client id).

---

## 1. Identity Service  — auth, accounts, profile

Owns the `users`, `tokens` tables. Issues + validates JWTs for every other service.

| Method | Path | Auth | Purpose | Frontend |
|---|---|---|---|---|
| POST | `/users` | — | Register | `auth.service.register` |
| POST | `/tokens/authentication` | — | Login → `{ user, token }` | `auth.service.login` |
| POST | `/tokens/password-reset` | — | Send reset link | `auth.service.forgotPassword` |
| PUT | `/users/password` | — (reset token) | Reset password via token | `auth.service.resetPassword` |
| GET | `/me` | 🔒 | Current user (wire real auth-context) | `auth-provider` (DEMO_USER swap) |
| PATCH | `/me` | 🔒👤 | Update name / headline / bio / avatar | profile, settings (account tab) |
| PUT | `/me/password` | 🔒👤 | Change password (knows current pw) | profile (security), settings |
| GET | `/me/settings` | 🔒👤 | Preferences (language, theme, notif toggles) | settings |
| PATCH | `/me/settings` | 🔒👤 | Save preferences | settings (notifications tab) |
| DELETE | `/me` | 🔒👤 | Delete account | settings (optional) |

> Optional (book-style): `PUT /users/activated` for email activation if you add it.

---

## 2. Catalog Service — courses, categories, instructors, curriculum, reviews

Owns `courses`, `modules`, `lessons`, `categories`, `instructors`, `reviews`.
Read-heavy + public/SEO. Curriculum writes are instructor-scoped.

### Courses — public reads
| Method | Path | Auth | Purpose | Frontend |
|---|---|---|---|---|
| GET | `/courses` | — | List + filters `?search&category&sort&page&pageSize&ids&instructorId` | catalog, popular, wishlist, cart |
| GET | `/courses/:slug` | — | Course detail (with modules, reviews) | course detail, `generateMetadata`, JSON-LD, sitemap |
| GET | `/courses/:id` | — | By numeric id (learn page, edit) | `courses.service.getById` |

### Courses — instructor writes (Studio)
| Method | Path | Auth | Purpose | Frontend |
|---|---|---|---|---|
| POST | `/courses` | 🔒 | Create course (nested modules/lessons) | studio → new course |
| PATCH | `/courses/:id` | 🔒🎓 | Update course fields | studio → edit |
| DELETE | `/courses/:id` | 🔒🎓 | Delete course | studio → courses list |
| PUT | `/courses/:id/publish` | 🔒🎓 | Publish / unpublish (`{ isPublished }`) | studio |

### Curriculum (if you want granular editing vs. nested-on-course)
| Method | Path | Auth |
|---|---|---|
| POST/PATCH/DELETE | `/courses/:id/modules[/:moduleId]` | 🔒🎓 |
| POST/PATCH/DELETE | `/modules/:id/lessons[/:lessonId]` | 🔒🎓 |
| PUT | `/courses/:id/curriculum/reorder` | 🔒🎓 |

### Categories & Instructors
| Method | Path | Auth | Purpose | Frontend |
|---|---|---|---|---|
| GET | `/categories` | — | List categories | categories page, course form |
| POST | `/categories` | ⚙️ | Create | admin categories |
| PATCH / DELETE | `/categories/:id` | ⚙️ | Edit / remove | admin categories |
| GET | `/instructors` | — | List instructors | instructors |
| GET | `/instructors/:id` | — | Instructor profile + their courses | instructor detail |

### Reviews
| Method | Path | Auth | Purpose |
|---|---|---|---|
| GET | `/courses/:id/reviews` | — | List reviews (paginated) |
| POST | `/courses/:id/reviews` | 🔒 | Add review (must be enrolled) |
| DELETE | `/reviews/:id` | 🔒👤 | Remove own review |

---

## 3. Learning Service — enrollment, progress, certificates, quizzes

Owns `enrollments`, `lesson_progress`, `certificates`, `quizzes`, `quiz_attempts`.
Needs course/lesson identity from Catalog (via API call or replicated ids/events).

| Method | Path | Auth | Purpose | Frontend |
|---|---|---|---|---|
| POST | `/courses/:id/enroll` | 🔒 | Enroll (free instantly; paid after order paid) | course detail, checkout success |
| GET | `/me/courses` | 🔒 | Enrolled courses + progress | dashboard, my-courses |
| GET | `/me/stats` | 🔒 | Dashboard stats (hours, streak, counts) | dashboard, profile |
| GET | `/me/courses/:id/progress` | 🔒 | Progress for a course | learn |
| PUT | `/me/lessons/:id/progress` | 🔒 | Mark lesson complete/incomplete (`{ completed }`) | learn (mark complete) |
| GET | `/me/certificates` | 🔒 | Earned certificates | certificates, dashboard |
| GET | `/quizzes/:id` | 🔒 | Quiz (questions **without** correct answers) | quiz |
| POST | `/quizzes/:id/attempts` | 🔒 | Submit answers → `{ score, passed }` | quiz (submit) |
| GET | `/me/quizzes/:id/attempts` | 🔒 | Attempt history | quiz (score history) |

> Security note: today the quiz score is computed **client-side** (`correctIndex` is
> shipped to the browser). Move grading server-side — `GET /quizzes/:id` must omit
> the correct answers, and `POST .../attempts` computes the score.

---

## 4. Commerce Service — cart, orders, payments

Owns `orders`, `order_items`, optionally `carts`. Emits "order.paid" → Learning enrolls the user.

| Method | Path | Auth | Purpose | Frontend |
|---|---|---|---|---|
| GET | `/me/cart` | 🔒 | Server cart (currently localStorage — optional to move) | cart |
| POST | `/me/cart/items` | 🔒 | Add course/lesson (`{ courseId, lessonId? }`) | course detail, cart |
| DELETE | `/me/cart/items/:key` | 🔒 | Remove line | cart |
| DELETE | `/me/cart` | 🔒 | Clear cart | checkout |
| POST | `/orders` | 🔒 | Checkout: create order from cart (`{ items, paymentMethod }`) → payment intent | checkout |
| GET | `/me/orders` | 🔒 | Purchase history | purchases |
| GET | `/me/orders/:id` | 🔒👤 | Single order | purchases detail |
| POST | `/payments/webhook` | — (signed) | Payment provider callback → mark order paid | — |

> `order_items` holds exactly one of `course_id` / `lesson_id` (matches `CartEntry`).

---

## 5. Engagement Service — notifications, wishlist

Owns `notifications`, `wishlists`. Low-risk to keep merged with Identity early on.

| Method | Path | Auth | Purpose | Frontend |
|---|---|---|---|---|
| GET | `/me/notifications` | 🔒 | List notifications | notifications, navbar badge |
| POST | `/me/notifications/read-all` | 🔒 | Mark all read | notifications |
| PATCH | `/me/notifications/:id/read` | 🔒 | Mark one read | notifications |
| GET | `/me/wishlist` | 🔒 | Wishlist course ids (currently localStorage) | wishlist |
| POST | `/me/wishlist/:courseId` | 🔒 | Add | course card / detail |
| DELETE | `/me/wishlist/:courseId` | 🔒 | Remove | wishlist, course card |

---

## 6. Admin & Analytics — back-office + instructor insights

Cross-cutting; in a microservice split this becomes an aggregating gateway that
fans out to the services above (or each service exposes its own `/admin/*`).

### Platform admin
| Method | Path | Auth | Purpose | Frontend |
|---|---|---|---|---|
| GET | `/admin/stats` | ⚙️ | Platform KPIs | admin overview |
| GET | `/admin/users` | ⚙️ | All users | admin users |
| PATCH | `/admin/users/:id` | ⚙️ | Suspend / activate (`{ status }`) | admin users |
| DELETE | `/admin/users/:id` | ⚙️ | Remove user | admin users |
| GET | `/admin/courses` | ⚙️ | All courses (incl. unpublished) | admin courses |
| PATCH | `/admin/courses/:id` | ⚙️ | Approve / unpublish | admin courses |
| DELETE | `/admin/courses/:id` | ⚙️ | Remove course | admin courses |

### Instructor analytics (Studio)
| Method | Path | Auth | Purpose | Frontend |
|---|---|---|---|---|
| GET | `/me/teaching/stats` | 🔒🎓 | Revenue, students, ratings overview | studio dashboard, analytics |
| GET | `/me/courses/:id/analytics` | 🔒🎓 | Per-course enrollment/revenue | studio analytics |

---

## Microservice split (recommended path)

```
                    ┌─────────────────┐
   client  ───────► │   API Gateway   │  (routing, JWT verify, rate-limit)
                    └───────┬─────────┘
      ┌──────────────┬──────┼───────────────┬────────────────┐
      ▼              ▼      ▼                ▼                ▼
 ┌─────────┐  ┌───────────┐  ┌────────────┐  ┌───────────┐  ┌──────────────┐
 │Identity │  │  Catalog  │  │  Learning  │  │ Commerce  │  │ Engagement   │
 │ users   │  │ courses   │  │ enroll     │  │ orders    │  │ notifications│
 │ tokens  │  │ categories│  │ progress   │  │ cart      │  │ wishlist     │
 │ profile │  │ reviews   │  │ certs/quiz │  │ payments  │  │              │
 └─────────┘  └───────────┘  └────────────┘  └───────────┘  └──────────────┘
```

**Cross-service dependencies to plan for:**
- Everyone verifies JWTs issued by **Identity** (shared public key / introspection).
- **Learning** & **Commerce** need course/lesson price + identity from **Catalog** →
  call Catalog's API, or subscribe to `course.updated` events and cache locally.
- **Commerce** emits `order.paid` → **Learning** enrolls the buyer (async, event-driven).
- **Learning** emits `course.completed` → **Engagement** sends a notification + issues cert.
- **Admin/Analytics** aggregates read-only data from all services (BFF/gateway).

**Pragmatic start:** build it as a **modular monolith** — one Go service, packages
per bounded context above (`internal/identity`, `internal/catalog`, …), each owning its
tables and exposing an interface. Split a package into its own service only when a real
scaling/ownership reason appears. The URL scheme above stays identical either way.
