package uz.chashma.lms.enrollment;

import org.springframework.jdbc.core.RowCallbackHandler;
import org.springframework.jdbc.core.RowMapper;
import org.springframework.jdbc.core.simple.JdbcClient;
import org.springframework.stereotype.Repository;
import uz.chashma.lms.shared.NotFoundException;

import java.time.OffsetDateTime;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.Set;

/** Go enrollment-service/internal/data/enrollments.go porti. */
@Repository
class EnrollmentRepository {

    private final JdbcClient jdbc;

    EnrollmentRepository(JdbcClient jdbc) {
        this.jdbc = jdbc;
    }

    private static final RowMapper<EnrollmentDto> MAPPER = (rs, rowNum) -> {
        EnrollmentDto e = new EnrollmentDto();
        e.id = rs.getLong("id");
        e.createdAt = rs.getObject("created_at", OffsetDateTime.class).toInstant();
        e.userId = rs.getLong("user_id");
        e.courseId = rs.getLong("course_id");
        return e;
    };

    record InsertResult(EnrollmentDto enrollment, boolean isNew) {
    }

    /** Idempotent: allaqachon yozilgan bo'lsa mavjud yozuvni qaytaradi. */
    InsertResult insert(long userId, long courseId) {
        EnrollmentDto enrollment = new EnrollmentDto();
        enrollment.userId = userId;
        enrollment.courseId = courseId;

        var inserted = new boolean[1];
        jdbc.sql("""
                INSERT INTO enrollment.enrollments (user_id, course_id)
                VALUES (:userId, :courseId)
                ON CONFLICT (user_id, course_id) DO NOTHING
                RETURNING id, created_at
                """)
                .param("userId", userId)
                .param("courseId", courseId)
                .query((RowCallbackHandler) rs -> {
                    enrollment.id = rs.getLong("id");
                    enrollment.createdAt = rs.getObject("created_at", OffsetDateTime.class).toInstant();
                    inserted[0] = true;
                });

        if (inserted[0]) {
            return new InsertResult(enrollment, true);
        }

        // Konflikt — mavjud yozuvni o'qiymiz.
        jdbc.sql("SELECT id, created_at FROM enrollment.enrollments WHERE user_id = :userId AND course_id = :courseId")
                .param("userId", userId)
                .param("courseId", courseId)
                .query((RowCallbackHandler) rs -> {
                    enrollment.id = rs.getLong("id");
                    enrollment.createdAt = rs.getObject("created_at", OffsetDateTime.class).toInstant();
                });

        return new InsertResult(enrollment, false);
    }

    Optional<EnrollmentDto> findById(long id) {
        if (id < 1) {
            return Optional.empty();
        }
        return jdbc.sql("SELECT id, created_at, user_id, course_id FROM enrollment.enrollments WHERE id = :id")
                .param("id", id)
                .query(MAPPER)
                .optional();
    }

    List<EnrollmentDto> listByUser(long userId) {
        return jdbc.sql("""
                SELECT id, created_at, user_id, course_id
                FROM enrollment.enrollments
                WHERE user_id = :userId
                ORDER BY created_at DESC, id DESC
                """)
                .param("userId", userId)
                .query(MAPPER)
                .list();
    }

    record EnrollmentPage(List<EnrollmentDto> enrollments, int total) {
    }

    /** /me/courses uchun sahifalangan variant (stats hisoblari to'liq ro'yxatni ishlatadi). */
    EnrollmentPage pageByUser(long userId, int page, int pageSize) {
        var totalHolder = new int[1];
        List<EnrollmentDto> items = jdbc.sql("""
                SELECT count(*) OVER() AS total, id, created_at, user_id, course_id
                FROM enrollment.enrollments
                WHERE user_id = :userId
                ORDER BY created_at DESC, id DESC
                LIMIT :limit OFFSET :offset
                """)
                .param("userId", userId)
                .param("limit", pageSize)
                .param("offset", (page - 1) * pageSize)
                .query((rs, rowNum) -> {
                    totalHolder[0] = rs.getInt("total");
                    return MAPPER.mapRow(rs, rowNum);
                })
                .list();
        return new EnrollmentPage(items, totalHolder[0]);
    }

    /**
     * Darslarga kirish beradi (kurs sotib olinganda barcha darslar, alohida
     * dars sotib olinganda bittasi).
     */
    void grantLessonAccess(long userId, long courseId, List<Long> lessonIds) {
        for (Long lessonId : lessonIds) {
            jdbc.sql("""
                    INSERT INTO enrollment.lesson_access (user_id, lesson_id, course_id)
                    VALUES (:userId, :lessonId, :courseId)
                    ON CONFLICT (user_id, lesson_id) DO NOTHING
                    """)
                    .param("userId", userId)
                    .param("lessonId", lessonId)
                    .param("courseId", courseId)
                    .update();
        }
    }

    /** Dars tugatildi/tugatilmadi belgisini qo'yadi. */
    void setLessonCompleted(long userId, long lessonId, boolean completed) {
        int updated = jdbc.sql("""
                UPDATE enrollment.lesson_access
                SET completed_at = CASE WHEN :completed THEN NOW() ELSE NULL END
                WHERE user_id = :userId AND lesson_id = :lessonId
                """)
                .param("completed", completed)
                .param("userId", userId)
                .param("lessonId", lessonId)
                .update();
        if (updated == 0) {
            throw new NotFoundException();
        }
    }

    /** Har bir kurs bo'yicha tugatilgan darslar soni. */
    Map<Long, Integer> completedCounts(long userId) {
        Map<Long, Integer> counts = new HashMap<>();
        jdbc.sql("""
                SELECT course_id, count(*) AS n
                FROM enrollment.lesson_access
                WHERE user_id = :userId AND completed_at IS NOT NULL
                GROUP BY course_id
                """)
                .param("userId", userId)
                .query((RowCallbackHandler) rs -> counts.put(rs.getLong("course_id"), rs.getInt("n")));
        return counts;
    }

    /** Tugatilgan dars id'lari (currentLesson uchun). */
    Set<Long> completedLessonIds(long userId) {
        Set<Long> done = new HashSet<>();
        jdbc.sql("SELECT lesson_id FROM enrollment.lesson_access WHERE user_id = :userId AND completed_at IS NOT NULL")
                .param("userId", userId)
                .query((RowCallbackHandler) rs -> done.add(rs.getLong("lesson_id")));
        return done;
    }

    /** Checkout: foydalanuvchi allaqachon yozilgan kurslar. */
    Set<Long> ownedCourses(long userId, List<Long> courseIds) {
        Set<Long> owned = new HashSet<>();
        if (courseIds.isEmpty()) {
            return owned;
        }
        jdbc.sql("SELECT course_id FROM enrollment.enrollments WHERE user_id = :userId AND course_id IN (:ids)")
                .param("userId", userId)
                .param("ids", courseIds)
                .query((RowCallbackHandler) rs -> owned.add(rs.getLong("course_id")));
        return owned;
    }

    /** Checkout: foydalanuvchida kirish huquqi bor darslar. */
    Set<Long> ownedLessons(long userId, List<Long> lessonIds) {
        Set<Long> owned = new HashSet<>();
        if (lessonIds.isEmpty()) {
            return owned;
        }
        jdbc.sql("SELECT lesson_id FROM enrollment.lesson_access WHERE user_id = :userId AND lesson_id IN (:ids)")
                .param("userId", userId)
                .param("ids", lessonIds)
                .query((RowCallbackHandler) rs -> owned.add(rs.getLong("lesson_id")));
        return owned;
    }

    /** Paywall: foydalanuvchining kursda kirish huquqi bor darslari. */
    List<Long> accessibleLessonIds(long userId, long courseId) {
        return jdbc.sql("SELECT lesson_id FROM enrollment.lesson_access WHERE user_id = :userId AND course_id = :courseId")
                .param("userId", userId)
                .param("courseId", courseId)
                .query(Long.class)
                .list();
    }

    /** Teaching stats: kurs -> yozilgan studentlar soni. */
    Map<Long, Integer> countsByCourses(List<Long> courseIds) {
        Map<Long, Integer> counts = new HashMap<>();
        if (courseIds.isEmpty()) {
            return counts;
        }
        jdbc.sql("""
                SELECT course_id, count(*) AS n
                FROM enrollment.enrollments
                WHERE course_id IN (:ids)
                GROUP BY course_id
                """)
                .param("ids", courseIds)
                .query((RowCallbackHandler) rs -> counts.put(rs.getLong("course_id"), rs.getInt("n")));
        return counts;
    }

    /** Kurslarga yozilgan noyob foydalanuvchilar soni. */
    int distinctStudentsForCourses(List<Long> courseIds) {
        if (courseIds.isEmpty()) {
            return 0;
        }
        return jdbc.sql("SELECT count(DISTINCT user_id) FROM enrollment.enrollments WHERE course_id IN (:ids)")
                .param("ids", courseIds)
                .query(Integer.class)
                .single();
    }

    record CompletedStats(Map<Long, Integer> counts, int activeStudents) {
    }

    /**
     * Teaching stats: kurs -> tugatilgan darslar yozuvlari soni (barcha
     * studentlar bo'yicha) va kamida bitta dars tugatgan noyob studentlar.
     */
    CompletedStats completedStatsByCourses(List<Long> courseIds) {
        Map<Long, Integer> counts = new HashMap<>();
        if (courseIds.isEmpty()) {
            return new CompletedStats(counts, 0);
        }

        jdbc.sql("""
                SELECT course_id, count(*) AS n
                FROM enrollment.lesson_access
                WHERE course_id IN (:ids) AND completed_at IS NOT NULL
                GROUP BY course_id
                """)
                .param("ids", courseIds)
                .query((RowCallbackHandler) rs -> counts.put(rs.getLong("course_id"), rs.getInt("n")));

        int active = jdbc.sql("""
                SELECT count(DISTINCT user_id)
                FROM enrollment.lesson_access
                WHERE course_id IN (:ids) AND completed_at IS NOT NULL
                """)
                .param("ids", courseIds)
                .query(Integer.class)
                .single();

        return new CompletedStats(counts, active);
    }

    /** Admin users ro'yxati: user -> yozilgan kurslari soni. */
    Map<Long, Integer> enrollmentCountsByUser(List<Long> ids) {
        Map<Long, Integer> counts = new HashMap<>();
        if (ids.isEmpty()) {
            return counts;
        }
        jdbc.sql("""
                SELECT user_id, count(*) AS n
                FROM enrollment.enrollments
                WHERE user_id IN (:ids)
                GROUP BY user_id
                """)
                .param("ids", ids)
                .query((RowCallbackHandler) rs -> counts.put(rs.getLong("user_id"), rs.getInt("n")));
        return counts;
    }
}
