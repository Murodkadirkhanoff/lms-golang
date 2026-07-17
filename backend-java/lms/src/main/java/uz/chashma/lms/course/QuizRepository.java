package uz.chashma.lms.course;

import org.springframework.dao.DataAccessException;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.jdbc.core.simple.JdbcClient;
import org.springframework.stereotype.Repository;
import org.springframework.transaction.annotation.Transactional;
import uz.chashma.lms.shared.PgErrors;

import java.sql.Array;
import java.sql.PreparedStatement;
import java.time.OffsetDateTime;
import java.util.List;
import java.util.Optional;

/** Go course-service/internal/data/quizzes.go porti. */
@Repository
class QuizRepository {

    private final JdbcClient jdbc;
    private final JdbcTemplate jdbcTemplate; // text[] parametrlari uchun

    QuizRepository(JdbcClient jdbc, JdbcTemplate jdbcTemplate) {
        this.jdbc = jdbc;
        this.jdbcTemplate = jdbcTemplate;
    }

    Optional<QuizDto> findByCourseId(long courseId) {
        if (courseId < 1) {
            return Optional.empty();
        }

        Optional<QuizDto> quiz = jdbc.sql("""
                SELECT id, course_id, title, passing_score, time_limit_minutes, version
                FROM course.quizzes
                WHERE course_id = :courseId
                """)
                .param("courseId", courseId)
                .query((rs, rowNum) -> {
                    QuizDto q = new QuizDto();
                    q.id = rs.getLong("id");
                    q.courseId = rs.getLong("course_id");
                    q.title = rs.getString("title");
                    q.passingScore = rs.getInt("passing_score");
                    q.timeLimitMinutes = rs.getInt("time_limit_minutes");
                    q.version = rs.getInt("version");
                    return q;
                })
                .optional();

        quiz.ifPresent(q -> q.questions = jdbc.sql("""
                SELECT id, question, options, correct_index, position
                FROM course.quiz_questions
                WHERE quiz_id = :quizId
                ORDER BY position, id
                """)
                .param("quizId", q.id)
                .query((rs, rowNum) -> {
                    QuizDto.QuestionDto question = new QuizDto.QuestionDto();
                    question.id = rs.getLong("id");
                    question.question = rs.getString("question");
                    Array options = rs.getArray("options");
                    question.options = List.of((String[]) options.getArray());
                    question.correctIndex = rs.getInt("correct_index");
                    question.position = rs.getInt("position");
                    return question;
                })
                .list());

        return quiz;
    }

    /** Kursning quizini savollari bilan butunlay almashtiradi. */
    @Transactional
    void upsert(QuizDto quiz) {
        try {
            jdbc.sql("""
                    INSERT INTO course.quizzes (course_id, title, passing_score, time_limit_minutes)
                    VALUES (:courseId, :title, :passingScore, :timeLimitMinutes)
                    ON CONFLICT (course_id)
                    DO UPDATE SET title = EXCLUDED.title, passing_score = EXCLUDED.passing_score,
                                  time_limit_minutes = EXCLUDED.time_limit_minutes,
                                  version = quizzes.version + 1
                    RETURNING id, version
                    """)
                    .param("courseId", quiz.courseId)
                    .param("title", quiz.title)
                    .param("passingScore", quiz.passingScore)
                    .param("timeLimitMinutes", quiz.timeLimitMinutes)
                    .query((org.springframework.jdbc.core.RowCallbackHandler) rs -> {
                        quiz.id = rs.getLong("id");
                        quiz.version = rs.getInt("version");
                    });
        } catch (DataAccessException ex) {
            if ("quizzes_course_id_fkey".equals(PgErrors.constraint(ex))) {
                throw new CourseErrors.InvalidCourseException();
            }
            throw ex;
        }

        jdbc.sql("DELETE FROM course.quiz_questions WHERE quiz_id = :quizId")
                .param("quizId", quiz.id)
                .update();

        for (int i = 0; i < quiz.questions.size(); i++) {
            QuizDto.QuestionDto q = quiz.questions.get(i);
            int position = i;
            Long id = jdbcTemplate.query(con -> {
                PreparedStatement ps = con.prepareStatement("""
                        INSERT INTO course.quiz_questions (quiz_id, question, options, correct_index, position)
                        VALUES (?, ?, ?, ?, ?)
                        RETURNING id
                        """);
                ps.setLong(1, quiz.id);
                ps.setString(2, q.question);
                ps.setArray(3, con.createArrayOf("text", q.options.toArray(new String[0])));
                ps.setInt(4, q.correctIndex);
                ps.setInt(5, position);
                return ps;
            }, rs -> rs.next() ? rs.getLong(1) : 0L);
            q.id = id == null ? 0 : id;
        }
    }

    void insertAttempt(QuizAttemptDto attempt) {
        jdbc.sql("""
                INSERT INTO course.quiz_attempts (user_id, course_id, score)
                VALUES (:userId, :courseId, :score)
                RETURNING id, created_at
                """)
                .param("userId", attempt.userId)
                .param("courseId", attempt.courseId)
                .param("score", attempt.score)
                .query((org.springframework.jdbc.core.RowCallbackHandler) rs -> {
                    attempt.id = rs.getLong("id");
                    attempt.createdAt = rs.getObject("created_at", OffsetDateTime.class).toInstant();
                });
    }

    List<QuizAttemptDto> listAttempts(long userId, long courseId) {
        return jdbc.sql("""
                SELECT id, created_at, user_id, course_id, score
                FROM course.quiz_attempts
                WHERE user_id = :userId AND course_id = :courseId
                ORDER BY created_at DESC, id DESC
                LIMIT 20
                """)
                .param("userId", userId)
                .param("courseId", courseId)
                .query((rs, rowNum) -> {
                    QuizAttemptDto a = new QuizAttemptDto();
                    a.id = rs.getLong("id");
                    a.createdAt = rs.getObject("created_at", OffsetDateTime.class).toInstant();
                    a.userId = rs.getLong("user_id");
                    a.courseId = rs.getLong("course_id");
                    a.score = rs.getInt("score");
                    return a;
                })
                .list();
    }

    /** Kurslardagi barcha urinishlarning o'rtacha bali (bo'lmasa 0). */
    double avgScoreForCourses(List<Long> courseIds) {
        if (courseIds.isEmpty()) {
            return 0;
        }
        Double avg = jdbc.sql("""
                SELECT COALESCE(AVG(score), 0)
                FROM course.quiz_attempts
                WHERE course_id IN (:ids)
                """)
                .param("ids", courseIds)
                .query(Double.class)
                .single();
        return avg == null ? 0 : avg;
    }
}
