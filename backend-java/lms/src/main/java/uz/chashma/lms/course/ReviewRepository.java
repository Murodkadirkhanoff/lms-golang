package uz.chashma.lms.course;

import org.springframework.dao.DataAccessException;
import org.springframework.jdbc.core.RowCallbackHandler;
import org.springframework.jdbc.core.simple.JdbcClient;
import org.springframework.stereotype.Repository;
import uz.chashma.lms.course.api.ReviewDto;
import uz.chashma.lms.shared.PgErrors;
import uz.chashma.lms.shared.UiDefaults;

import java.time.OffsetDateTime;
import java.util.List;

/** Go course-service/internal/data/reviews.go porti. */
@Repository
class ReviewRepository {

    private final JdbcClient jdbc;

    ReviewRepository(JdbcClient jdbc) {
        this.jdbc = jdbc;
    }

    /** Bitta user bitta kursga bitta sharh — qayta yuborsa yangilanadi. */
    void upsert(ReviewDto review) {
        try {
            jdbc.sql("""
                    INSERT INTO course.reviews (course_id, user_id, user_name, rating, comment)
                    VALUES (:courseId, :userId, :userName, :rating, :comment)
                    ON CONFLICT (course_id, user_id)
                    DO UPDATE SET rating = EXCLUDED.rating, comment = EXCLUDED.comment,
                                  user_name = EXCLUDED.user_name, created_at = NOW()
                    RETURNING id, created_at
                    """)
                    .param("courseId", review.courseId)
                    .param("userId", review.userId)
                    .param("userName", review.user)
                    .param("rating", review.rating)
                    .param("comment", review.comment)
                    .query((RowCallbackHandler) rs -> {
                        review.id = rs.getLong("id");
                        review.createdAt = rs.getObject("created_at", OffsetDateTime.class).toInstant();
                    });
        } catch (DataAccessException ex) {
            if ("reviews_course_id_fkey".equals(PgErrors.constraint(ex))) {
                throw new CourseErrors.InvalidCourseException();
            }
            throw ex;
        }

        review.avatarColor = UiDefaults.avatarColor(review.userId);
    }

    List<ReviewDto> listForCourse(long courseId, int limit) {
        return jdbc.sql("""
                SELECT id, created_at, course_id, user_id, user_name, rating, comment
                FROM course.reviews
                WHERE course_id = :courseId
                ORDER BY created_at DESC, id DESC
                LIMIT :limit
                """)
                .param("courseId", courseId)
                .param("limit", limit)
                .query((rs, rowNum) -> {
                    ReviewDto r = new ReviewDto();
                    r.id = rs.getLong("id");
                    r.createdAt = rs.getObject("created_at", OffsetDateTime.class).toInstant();
                    r.courseId = rs.getLong("course_id");
                    r.userId = rs.getLong("user_id");
                    r.user = rs.getString("user_name");
                    r.rating = rs.getInt("rating");
                    r.comment = rs.getString("comment");
                    r.avatarColor = UiDefaults.avatarColor(r.userId);
                    return r;
                })
                .list();
    }
}
