# --- Build: JDK 25 + Maven wrapper ---
FROM eclipse-temurin:25-jdk AS build
WORKDIR /app

# Avval faqat pom + wrapper — dependencylar alohida layer'da cache'lanadi
COPY .mvn .mvn
COPY mvnw pom.xml ./
# Windows'da checkout qilinganda CRLF bo'lib qolishi mumkin
RUN sed -i 's/\r$//' mvnw && ./mvnw -q -B dependency:go-offline

COPY src src
RUN ./mvnw -q -B -DskipTests package \
    # Layered extract: dependencies / application alohida dirlarga —
    # runtime image'da har biri o'z layer'i bo'ladi (k8s pull'lari tezlashadi)
    && java -Djarmode=tools -jar target/lms-*.jar extract --layers --destination extracted

# --- Runtime: faqat JRE ---
FROM eclipse-temurin:25-jre AS run
WORKDIR /app

# curl — docker compose / k8s exec-healthcheck uchun (k8s httpGet probe'ga shart emas)
RUN apt-get update \
    && apt-get install -y --no-install-recommends curl \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --uid 10001 --create-home app \
    # Yuklangan fayllar shu yerga (compose'da volume sifatida mount qilinadi)
    && mkdir -p /app/uploads && chown app /app/uploads

COPY --from=build /app/extracted/dependencies/ ./
COPY --from=build /app/extracted/spring-boot-loader/ ./
COPY --from=build /app/extracted/snapshot-dependencies/ ./
COPY --from=build /app/extracted/application/ ./

USER app
EXPOSE 4000
# Konfiguratsiya to'liq env orqali (application.yml placeholderlari):
# LMS_PORT, LMS_DB_URL, LMS_DB_USER, LMS_DB_PASSWORD, JWT_SECRET, JWT_TTL,
# LMS_ENV, CORS_TRUSTED_ORIGINS. Health: GET /v1/healthcheck
ENTRYPOINT ["java", "-jar", "lms-0.0.1-SNAPSHOT.jar"]
