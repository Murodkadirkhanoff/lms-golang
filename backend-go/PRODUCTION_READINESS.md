# Production Readiness — backlog

Status of the Go backend against production requirements. Architecture and
functionality are production-grade and verified; the items below are the
remaining hardening tasks. Ordered by severity.

Legend — **Sev**: Critical / High / Medium / Low · **Est**: rough effort.

---

## P0 — Critical (must fix before production)

### 1. Large uploads fail due to server read timeout
- **Sev:** Critical · **Est:** 15m
- **Where:** `internal/app/app.go` (`http.Server{ReadTimeout: 15s, WriteTimeout: 30s}`),
  `internal/app/uploads.go` (`maxUploadBytes = 512MB`).
- **Problem:** `ReadTimeout` covers the entire request body. A 512 MB video
  upload cannot be read in 15 s (~34 MB/s required), so the connection is cut.
- **Fix:** Replace `ReadTimeout` with `ReadHeaderTimeout` (headers only), and/or
  serve the `/v1/uploads` route from a separate `http.Server`/handler with a
  longer/zero body timeout. Keep `WriteTimeout` off for the upload path.
- **Done when:** a 300 MB+ file uploads successfully end-to-end.

---

## P1 — High (observability; needed to operate in production)

### 2. No request logging / correlation IDs
- **Sev:** High · **Est:** 30m
- **Where:** `internal/app/router.go` (no access-log middleware).
- **Problem:** No per-request access logs or request IDs — debugging incidents
  and tracing a request across logs is not possible.
- **Fix:** Add `chi/middleware.RequestID`, `middleware.RealIP`, and a structured
  `slog` access-log middleware (method, path, status, duration, request id).
  Propagate the request id into `slog` records and error responses.
- **Done when:** every request logs one structured line with a correlation id.

### 3. No metrics / tracing
- **Sev:** High · **Est:** 1–2h
- **Problem:** No `/metrics` (Prometheus) or distributed tracing; no visibility
  into latency, error rate, throughput, pool saturation.
- **Fix:** Expose Prometheus metrics (HTTP histogram + pgx pool stats) on a
  separate admin port; optionally OpenTelemetry tracing around HTTP + DB.
- **Done when:** RED metrics and pool stats are scrapeable.

---

## P2 — Medium (security & ops hardening)

### 4. Category write endpoints are public
- **Sev:** Medium · **Est:** 10m
- **Where:** `internal/modules/courses/transport/routes.go` (categories POST/PATCH/DELETE outside the auth group).
- **Problem:** Anyone can create/edit/delete categories (a quirk carried over
  from the Java backend for parity).
- **Fix:** Move the three category-write routes into a
  `RequireRole(domain.RoleAdmin)` group.
- **Done when:** category writes return 401/403 without an admin token.

### 5. Liveness-only health; no readiness
- **Sev:** Medium · **Est:** 20m
- **Where:** `internal/app/health.go` (does not touch the DB).
- **Problem:** `/v1/healthcheck` returns OK even when Postgres is down; k8s
  cannot tell "process alive" from "ready to serve".
- **Fix:** Add a `/readyz` that runs `pool.Ping` with a short timeout; keep the
  current endpoint as liveness. Wire both into the compose/k8s probes.
- **Done when:** `/readyz` fails when the DB is unreachable.

### 6. Rate limiter is per-instance (in-memory)
- **Sev:** Medium · **Est:** 1–2h
- **Where:** `internal/platform/web/ratelimit.go`.
- **Problem:** Behind N replicas the effective limit scales ×N (each pod counts
  independently).
- **Fix:** Back the limiter with a shared store (Redis) when running >1 replica,
  or enforce the limit at the gateway/ingress.
- **Done when:** the limit holds regardless of replica count.

### 7. Uploads validated by extension only; local-disk storage
- **Sev:** Medium · **Est:** 1h (sniff) / larger (S3)
- **Where:** `internal/app/uploads.go`.
- **Problem:** Only the filename extension is checked (no content sniffing);
  files live on a local volume (single node, no CDN/durability).
- **Fix:** Sniff the content type (`http.DetectContentType`) and reject
  mismatches; set correct `Content-Type` on static serving. For scale/durability
  move to S3-compatible storage (only `uploads.go` changes; URL contract stays).
- **Done when:** disguised files are rejected; object storage is pluggable.

### 8. No server-side query timeout
- **Sev:** Medium · **Est:** 20m
- **Problem:** Handlers use `r.Context()` (cancels on client disconnect) but
  there is no upper bound on slow queries.
- **Fix:** Add a per-request timeout middleware (e.g. 10–15 s) or wrap DB calls
  with `context.WithTimeout`; set a Postgres `statement_timeout`.
- **Done when:** a runaway query is cancelled server-side.

---

## P3 — Low (quality / nice to have)

### 9. Only domain unit tests
- **Sev:** Low · **Est:** half day
- **Problem:** No automated HTTP-handler or integration tests (flows were smoke-
  tested manually).
- **Fix:** Add handler tests with `httptest` + fakes for the contract ports, and
  an integration suite against a throwaway Postgres (testcontainers) covering
  register→enroll→progress→certificate and checkout.
- **Done when:** CI runs unit + integration on every push.

### 10. Config / secrets & TLS
- **Sev:** Low · **Est:** varies
- **Notes:** `JWT_SECRET` comes from env (fine); add secret management + rotation
  for real deployments. TLS is expected to terminate at a load balancer/ingress
  (no in-app HTTPS) — document/confirm that assumption.

---

## Suggested order

1. **#1** (upload bug) — correctness blocker.
2. **#2, #3** — you cannot operate blind.
3. **#4, #5, #8** — quick security/ops wins.
4. **#6, #7** — needed when you scale horizontally / need durable media.
5. **#9, #10** — ongoing quality.
