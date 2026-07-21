# LMS — Go modular monolith

A production-grade Go rewrite of the `backend-java/lms` Spring Boot application.
DDD-inspired modular architecture: one deployable today, three independently
extractable microservices tomorrow. See [MIGRATION_PLAN.md](MIGRATION_PLAN.md)
for the full design rationale, and [PRODUCTION_READINESS.md](PRODUCTION_READINESS.md)
for the remaining hardening backlog before production.

## Layout

```
cmd/api/                        entrypoint
internal/
  app/                          composition root: config load, migrations, router, uploads, health
  platform/                     framework kernel (NO business logic)
    config/                     env configuration
    database/                   pgx pool + embedded-migration runner + pg error mapping
    web/                        JSON I/O, error envelope, validator, JWT auth, rate limit, CORS
  modules/
    users/                      auth, profile, admin, password reset   (owns `auth` schema)
    courses/                    catalog, categories, lessons, reviews, quizzes, instructors, Q&A (owns `course` schema)
    enrollment/                 enroll, progress, orders/checkout, certificates, notifications, teaching (owns `enrollment` schema)
```

Each module is a self-contained bounded context:

```
modules/<ctx>/
  domain/          entities, rules, errors, value helpers
  application/     use-case services + ports (interfaces)
  infrastructure/  pgx repositories (hand-written SQL, no ORM)
  transport/       chi HTTP handlers + routes
  contract/        public port: interface(s) + plain structs peers may import
  migrations/      //go:embed *.sql owned by this context
  module.go        the module's composition root
```

## Rules enforced

- **No `shared/`/`common/`/`utils/`/`helpers/` and no shared business code.** The
  only cross-cutting package is `platform`, which is framework plumbing (like
  chi or pgx) with zero domain knowledge.
- **A module may import only another module's `contract/` package** — never its
  `domain`/`application`/`infrastructure`/`transport`. Verified:

  ```
  users      → courses/contract, enrollment/contract   (admin dashboard)
  courses    → users/contract, enrollment/contract      (instructor names, review gate, paywall)
  enrollment → courses/contract, users/contract          (catalog, names)
  ```
- **Database-per-module thinking.** Each module embeds its own migrations and
  versions them under `schema_migrations_<ctx>`; its repositories touch only its
  own schema. Cross-context references (e.g. `course.instructor_id`) carry no
  cross-schema FK by design.

The contract *implementations* form a DAG (`users → courses → enrollment`), so
there is no construction or import cycle even though HTTP handlers call across
contexts.

## Run

```bash
# 1. Full stack (Postgres + backend + mailpit + frontend) from the repo root
export JWT_SECRET=$(openssl rand -hex 32)
docker compose up --build          # uses ../docker-compose.yml

# 2. or this module locally against your own Postgres
cp .env.example .env               # fill JWT_SECRET
export $(grep -v '^#' .env | xargs)
go run ./cmd/api
```

The single source of truth for containerised runs is the repo-root
`docker-compose.yml`; this module has no compose file of its own.

Migrations run automatically on startup. Health check: `GET /v1/healthcheck`.

## Test

```bash
go build ./...
go vet ./...
go test ./...
```

## API surface (v1)

| Method | Path | Auth | Module |
|--------|------|------|--------|
| POST | /v1/users | public* | users |
| POST | /v1/tokens/authentication | public* | users |
| POST | /v1/tokens/password-reset | public* | users |
| PUT  | /v1/users/password | public* | users |
| GET/PUT | /v1/me, /v1/me/profile, /v1/me/password | auth | users |
| GET/PATCH | /v1/admin/users, /v1/admin/users/{id}/role, /v1/admin/stats | admin | users |
| GET | /v1/courses, /v1/courses/{id} | public | courses |
| POST/PATCH/DELETE | /v1/courses, /v1/courses/{id} | auth (owner/admin) | courses |
| GET/POST/PATCH/DELETE | /v1/categories… | public | courses |
| GET | /v1/instructors, /v1/instructors/{id} | public | courses |
| GET | /v1/quizzes/{id}, /v1/lessons/{id}/questions | public | courses |
| PUT | /v1/courses/{id}/quiz | auth (owner/admin) | courses |
| POST | /v1/courses/{id}/reviews | auth (enrolled) | courses |
| POST/GET | /v1/quizzes/{id}/attempts | auth | courses |
| POST | /v1/lessons/{id}/questions | auth | courses |
| POST | /v1/courses/{id}/enroll | auth | enrollment |
| PATCH | /v1/enrollments/{id}/progress | auth | enrollment |
| GET | /v1/me/stats, /courses, /certificates, /notifications, /orders, /teaching/stats | auth | enrollment |
| POST | /v1/me/orders (checkout) | auth | enrollment |
| GET | /v1/me/certificates/{id}/download (PDF) | auth | enrollment |
| POST | /v1/uploads | auth | app (media) |

\* rate-limited per IP.

## Extracting a module into a service

Move `internal/modules/<ctx>/**` plus the `platform` kernel into a new `cmd`,
point its embedded migrations at a dedicated database, and replace the injected
peer contracts (e.g. `users.Directory`, `enrollment.Gate`) with HTTP/gRPC
clients that satisfy the same interfaces. No code inside the module changes.
