package uz.chashma.lms.course;

import tools.jackson.databind.PropertyNamingStrategies;
import tools.jackson.databind.annotation.JsonNaming;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;
import uz.chashma.lms.course.api.CourseDto;
import uz.chashma.lms.shared.NotFoundException;
import uz.chashma.lms.shared.NotPermittedException;
import uz.chashma.lms.shared.Validator;
import uz.chashma.lms.shared.security.CurrentUser;
import uz.chashma.lms.shared.security.Roles;
import uz.chashma.lms.shared.security.UserPrincipal;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

import static uz.chashma.lms.shared.Validator.orEmpty;

/**
 * Go course-service cmd/api/quizzes.go porti. Frontend {id} sifatida KURS
 * id'sini yuboradi (learn sahifasi ROUTES.quiz(course.id)).
 */
@RestController
class QuizController {

    private final QuizRepository quizzes;
    private final CourseRepository courses;

    QuizController(QuizRepository quizzes, CourseRepository courses) {
        this.quizzes = quizzes;
        this.courses = courses;
    }

    // GET /v1/quizzes/{id} ({id} — kurs id).
    @GetMapping("/v1/quizzes/{id}")
    Map<String, Object> show(@PathVariable long id) {
        QuizDto quiz = quizzes.findByCourseId(id).orElseThrow(NotFoundException::new);
        return Map.of("quiz", quiz);
    }

    @JsonNaming(PropertyNamingStrategies.SnakeCaseStrategy.class)
    record QuestionRequest(String question, List<String> options, int correctIndex) {
    }

    @JsonNaming(PropertyNamingStrategies.SnakeCaseStrategy.class)
    record UpsertQuizRequest(String title, int passingScore, int timeLimitMinutes,
                             List<QuestionRequest> questions) {
    }

    // PUT /v1/courses/{id}/quiz (kurs egasi yoki admin).
    @PutMapping("/v1/courses/{id}/quiz")
    Map<String, Object> upsert(@PathVariable long id, @RequestBody UpsertQuizRequest input) {
        CourseDto course = courses.findByIdOrSlug(Long.toString(id)).orElseThrow(NotFoundException::new);

        if (!canModifyCourse(course)) {
            throw new NotPermittedException();
        }

        QuizDto quiz = new QuizDto();
        quiz.courseId = id;
        quiz.title = orEmpty(input.title());
        quiz.passingScore = input.passingScore() == 0 ? 70 : input.passingScore();
        quiz.timeLimitMinutes = input.timeLimitMinutes() == 0 ? 10 : input.timeLimitMinutes();

        if (input.questions() != null) {
            for (QuestionRequest q : input.questions()) {
                QuizDto.QuestionDto question = new QuizDto.QuestionDto();
                question.question = orEmpty(q.question());
                question.options = q.options() == null ? new ArrayList<>() : q.options();
                question.correctIndex = q.correctIndex();
                quiz.questions.add(question);
            }
        }

        Validator v = new Validator();
        validateQuiz(v, quiz);
        v.throwIfInvalid();

        try {
            quizzes.upsert(quiz);
        } catch (CourseErrors.InvalidCourseException e) {
            throw new NotFoundException();
        }

        return Map.of("quiz", quiz);
    }

    record SubmitAttemptRequest(int score) {
    }

    // POST /v1/quizzes/{id}/attempts. {id} — kurs id. Baholash clientda,
    // server natijani tarix va analitika uchun saqlaydi.
    @PostMapping("/v1/quizzes/{id}/attempts")
    ResponseEntity<Map<String, Object>> submitAttempt(@PathVariable long id,
                                                      @RequestBody SubmitAttemptRequest input) {
        // Quiz mavjudligini tekshiramiz — bo'lmagan kursga urinish yozilmaydi.
        quizzes.findByCourseId(id).orElseThrow(NotFoundException::new);

        Validator v = new Validator();
        v.check(input.score() >= 0 && input.score() <= 100, "score", "must be between 0 and 100");
        v.throwIfInvalid();

        UserPrincipal claims = CurrentUser.get();

        QuizAttemptDto attempt = new QuizAttemptDto();
        attempt.userId = claims.id();
        attempt.courseId = id;
        attempt.score = input.score();

        quizzes.insertAttempt(attempt);

        return ResponseEntity.status(HttpStatus.CREATED).body(Map.of("attempt", attempt));
    }

    // GET /v1/quizzes/{id}/attempts — o'z urinishlari (score history).
    @GetMapping("/v1/quizzes/{id}/attempts")
    Map<String, Object> listAttempts(@PathVariable long id) {
        UserPrincipal claims = CurrentUser.get();
        return Map.of("attempts", quizzes.listAttempts(claims.id(), id));
    }

    private static boolean canModifyCourse(CourseDto course) {
        UserPrincipal claims = CurrentUser.get();
        if (claims == null) {
            return false;
        }
        return claims.id() == course.instructorId || Roles.ADMIN.equals(claims.role());
    }

    // Go data.ValidateQuiz bilan bir xil xabarlar.
    private static void validateQuiz(Validator v, QuizDto quiz) {
        v.check(!quiz.title.isEmpty(), "title", "must be provided");
        v.check(quiz.passingScore >= 0 && quiz.passingScore <= 100, "passing_score", "must be between 0 and 100");
        v.check(quiz.timeLimitMinutes > 0, "time_limit_minutes", "must be greater than zero");
        v.check(!quiz.questions.isEmpty(), "questions", "must contain at least one question");

        for (int i = 0; i < quiz.questions.size(); i++) {
            QuizDto.QuestionDto q = quiz.questions.get(i);
            String key = "questions[%d]".formatted(i);
            v.check(!q.question.isEmpty(), key + ".question", "must be provided");
            v.check(q.options.size() >= 2, key + ".options", "must have at least 2 options");
            v.check(q.correctIndex >= 0 && q.correctIndex < q.options.size(),
                    key + ".correct_index", "must point to an option");
        }
    }
}
