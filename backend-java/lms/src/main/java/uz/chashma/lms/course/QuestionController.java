package uz.chashma.lms.course;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.jdbc.core.simple.JdbcClient;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;
import uz.chashma.lms.auth.api.UserApi;
import uz.chashma.lms.shared.NotFoundException;
import uz.chashma.lms.shared.Validator;
import uz.chashma.lms.shared.security.CurrentUser;
import uz.chashma.lms.shared.security.UserPrincipal;

import java.time.Instant;
import java.time.OffsetDateTime;
import java.util.List;
import java.util.Map;

import static uz.chashma.lms.shared.Validator.byteLength;
import static uz.chashma.lms.shared.Validator.orEmpty;

/** Learn sahifasi Q&A tabi: dars bo'yicha savollar (javob threadi keyinroq). */
@RestController
class QuestionController {

    /** Frontend LessonQuestion tipi bilan bir xil JSON. */
    record QuestionDto(long id, Instant createdAt, String user, String question) {
    }

    private final JdbcClient jdbc;
    private final UserApi userApi;

    QuestionController(JdbcClient jdbc, UserApi userApi) {
        this.jdbc = jdbc;
        this.userApi = userApi;
    }

    // GET /v1/lessons/{id}/questions — so'nggi 100 ta savol.
    @GetMapping("/v1/lessons/{id}/questions")
    Map<String, Object> list(@PathVariable long id) {
        List<QuestionDto> items = jdbc.sql("""
                SELECT id, created_at, user_name, question
                FROM course.lesson_questions
                WHERE lesson_id = :lessonId
                ORDER BY created_at DESC, id DESC
                LIMIT 100
                """)
                .param("lessonId", id)
                .query((rs, rowNum) -> new QuestionDto(
                        rs.getLong("id"),
                        rs.getObject("created_at", OffsetDateTime.class).toInstant(),
                        rs.getString("user_name"),
                        rs.getString("question")))
                .list();
        return Map.of("items", items);
    }

    record AskRequest(String question) {
    }

    // POST /v1/lessons/{id}/questions (auth).
    @PostMapping("/v1/lessons/{id}/questions")
    ResponseEntity<Map<String, Object>> ask(@PathVariable long id, @RequestBody AskRequest input) {
        UserPrincipal claims = CurrentUser.get();
        String question = orEmpty(input.question()).trim();

        Validator v = new Validator();
        v.check(!question.isEmpty(), "question", "must be provided");
        v.check(byteLength(question) <= 2000, "question", "must not be more than 2000 bytes long");
        v.throwIfInvalid();

        String userName = "Student";
        List<UserApi.UserSummary> users = userApi.findByIds(List.of(claims.id()));
        if (!users.isEmpty()) {
            userName = users.get(0).name();
        }

        QuestionDto saved;
        try {
            saved = jdbc.sql("""
                    INSERT INTO course.lesson_questions (lesson_id, user_id, user_name, question)
                    VALUES (:lessonId, :userId, :userName, :question)
                    RETURNING id, created_at
                    """)
                    .param("lessonId", id)
                    .param("userId", claims.id())
                    .param("userName", userName)
                    .param("question", question)
                    .query((rs, rowNum) -> new QuestionDto(
                            rs.getLong("id"),
                            rs.getObject("created_at", OffsetDateTime.class).toInstant(),
                            null, null))
                    .single();
        } catch (org.springframework.dao.DataAccessException e) {
            // FK buzilishi — dars mavjud emas.
            throw new NotFoundException();
        }

        return ResponseEntity.status(HttpStatus.CREATED)
                .body(Map.of("question", new QuestionDto(saved.id(), saved.createdAt(), userName, question)));
    }
}
