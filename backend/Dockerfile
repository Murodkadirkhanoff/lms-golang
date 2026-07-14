# Uchala Go servis uchun umumiy Dockerfile.
# Qaysi servis qurilishi build-arg bilan tanlanadi:
#   docker build --build-arg SERVICE=auth-service .
FROM golang:1.26-alpine AS build
ARG SERVICE
WORKDIR /src

# Avval modul fayllari — dependency layer keshda qoladi.
COPY pkg/ ./pkg/
COPY ${SERVICE}/go.mod ${SERVICE}/go.sum ./${SERVICE}/
WORKDIR /src/${SERVICE}
RUN go mod download

WORKDIR /src
COPY ${SERVICE}/ ./${SERVICE}/
WORKDIR /src/${SERVICE}
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/api ./cmd/api

FROM alpine:3.21
ARG SERVICE
RUN adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /bin/api ./api
# Servis startup'da shu papkadan migratsiyalarni o'zi qo'llaydi.
COPY --from=build /src/${SERVICE}/migrations ./migrations
USER app
ENTRYPOINT ["./api"]
