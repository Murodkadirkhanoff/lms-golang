package uz.chashma.lms.enrollment;

import com.fasterxml.jackson.annotation.JsonIgnore;
import org.springframework.jdbc.core.simple.JdbcClient;
import org.springframework.stereotype.Repository;
import uz.chashma.lms.shared.UiDefaults;

import java.time.Instant;
import java.time.OffsetDateTime;
import java.util.List;

/** Go enrollment-service/internal/data/certificates.go porti. */
@Repository
class CertificateRepository {

    /** Frontend Certificate tipi bilan bir xil JSON. */
    static class CertificateDto {
        public long id;
        public Instant issuedAt;
        @JsonIgnore
        public long userId;
        @JsonIgnore
        public long courseId;
        public String courseTitle;
        public String color;
    }

    private final JdbcClient jdbc;

    CertificateRepository(JdbcClient jdbc) {
        this.jdbc = jdbc;
    }

    /** Idempotent: allaqachon berilgan bo'lsa yangisini yaratmaydi. */
    boolean issue(long userId, long courseId, String courseTitle) {
        return jdbc.sql("""
                INSERT INTO enrollment.certificates (user_id, course_id, course_title)
                VALUES (:userId, :courseId, :courseTitle)
                ON CONFLICT (user_id, course_id) DO NOTHING
                RETURNING id
                """)
                .param("userId", userId)
                .param("courseId", courseId)
                .param("courseTitle", courseTitle)
                .query(Long.class)
                .optional()
                .isPresent();
    }

    List<CertificateDto> listByUser(long userId) {
        return jdbc.sql("""
                SELECT id, issued_at, user_id, course_id, course_title
                FROM enrollment.certificates
                WHERE user_id = :userId
                ORDER BY issued_at DESC, id DESC
                """)
                .param("userId", userId)
                .query((rs, rowNum) -> {
                    CertificateDto c = new CertificateDto();
                    c.id = rs.getLong("id");
                    c.issuedAt = rs.getObject("issued_at", OffsetDateTime.class).toInstant();
                    c.userId = rs.getLong("user_id");
                    c.courseId = rs.getLong("course_id");
                    c.courseTitle = rs.getString("course_title");
                    c.color = UiDefaults.thumbnailColor(c.courseId);
                    return c;
                })
                .list();
    }

    java.util.Optional<CertificateDto> findForUser(long id, long userId) {
        return jdbc.sql("""
                SELECT id, issued_at, user_id, course_id, course_title
                FROM enrollment.certificates
                WHERE id = :id AND user_id = :userId
                """)
                .param("id", id)
                .param("userId", userId)
                .query((rs, rowNum) -> {
                    CertificateDto c = new CertificateDto();
                    c.id = rs.getLong("id");
                    c.issuedAt = rs.getObject("issued_at", OffsetDateTime.class).toInstant();
                    c.userId = rs.getLong("user_id");
                    c.courseId = rs.getLong("course_id");
                    c.courseTitle = rs.getString("course_title");
                    c.color = UiDefaults.thumbnailColor(c.courseId);
                    return c;
                })
                .optional();
    }

    int countByUser(long userId) {
        return jdbc.sql("SELECT count(*) FROM enrollment.certificates WHERE user_id = :userId")
                .param("userId", userId)
                .query(Integer.class)
                .single();
    }
}
