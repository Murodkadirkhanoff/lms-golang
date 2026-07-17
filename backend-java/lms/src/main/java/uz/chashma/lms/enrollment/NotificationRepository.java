package uz.chashma.lms.enrollment;

import com.fasterxml.jackson.annotation.JsonIgnore;
import org.springframework.jdbc.core.simple.JdbcClient;
import org.springframework.stereotype.Repository;
import uz.chashma.lms.shared.NotFoundException;

import java.time.Instant;
import java.time.OffsetDateTime;
import java.util.List;

/** Go enrollment-service/internal/data/notifications.go porti. */
@Repository
class NotificationRepository {

    /** Frontend Notification tipi bilan bir xil JSON. */
    static class NotificationDto {
        public long id;
        public Instant createdAt;
        @JsonIgnore
        public long userId;
        public String type;
        public String title;
        public String body;
        public boolean read;
    }

    private final JdbcClient jdbc;

    NotificationRepository(JdbcClient jdbc) {
        this.jdbc = jdbc;
    }

    void insert(long userId, String type, String title, String body) {
        jdbc.sql("""
                INSERT INTO enrollment.notifications (user_id, type, title, body)
                VALUES (:userId, :type, :title, :body)
                """)
                .param("userId", userId)
                .param("type", type)
                .param("title", title)
                .param("body", body)
                .update();
    }

    List<NotificationDto> listByUser(long userId) {
        return jdbc.sql("""
                SELECT id, created_at, user_id, type, title, body, read
                FROM enrollment.notifications
                WHERE user_id = :userId
                ORDER BY created_at DESC, id DESC
                LIMIT 50
                """)
                .param("userId", userId)
                .query((rs, rowNum) -> {
                    NotificationDto n = new NotificationDto();
                    n.id = rs.getLong("id");
                    n.createdAt = rs.getObject("created_at", OffsetDateTime.class).toInstant();
                    n.userId = rs.getLong("user_id");
                    n.type = rs.getString("type");
                    n.title = rs.getString("title");
                    n.body = rs.getString("body");
                    n.read = rs.getBoolean("read");
                    return n;
                })
                .list();
    }

    void markAllRead(long userId) {
        jdbc.sql("UPDATE enrollment.notifications SET read = true WHERE user_id = :userId AND read = false")
                .param("userId", userId)
                .update();
    }

    /** Bitta bildirishnomani o'qildi qiladi (faqat egasiniki). */
    void markRead(long id, long userId) {
        int updated = jdbc.sql("UPDATE enrollment.notifications SET read = true WHERE id = :id AND user_id = :userId")
                .param("id", id)
                .param("userId", userId)
                .update();
        if (updated == 0) {
            throw new NotFoundException();
        }
    }
}
