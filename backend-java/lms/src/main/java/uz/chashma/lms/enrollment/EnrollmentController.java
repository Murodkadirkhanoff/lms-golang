package uz.chashma.lms.enrollment;

import tools.jackson.databind.PropertyNamingStrategies;
import tools.jackson.databind.annotation.JsonNaming;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PatchMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;
import uz.chashma.lms.course.api.CourseDto;
import uz.chashma.lms.course.api.CourseApi;
import uz.chashma.lms.shared.NotFoundException;
import uz.chashma.lms.shared.NotPermittedException;
import uz.chashma.lms.shared.Validator;
import uz.chashma.lms.shared.security.CurrentUser;
import uz.chashma.lms.shared.security.UserPrincipal;

import java.util.List;
import java.util.Map;

/** Go enrollment-service cmd/api/enrollments.go porti. */
@RestController
class EnrollmentController {

    private final EnrollmentRepository enrollments;
    private final EnrollmentService service;
    private final CourseApi courseApi;

    EnrollmentController(EnrollmentRepository enrollments, EnrollmentService service, CourseApi courseApi) {
        this.enrollments = enrollments;
        this.service = service;
        this.courseApi = courseApi;
    }

    // POST /v1/courses/{id}/enroll. Faqat bepul kurslar; pulliklar checkout orqali.
    @PostMapping("/v1/courses/{id}/enroll")
    ResponseEntity<Map<String, Object>> enroll(@PathVariable long id) {
        UserPrincipal claims = CurrentUser.get();

        List<CourseDto> courses = courseApi.coursesByIds(List.of(id));
        if (courses.isEmpty()) {
            throw new NotFoundException();
        }
        CourseDto course = courses.get(0);

        Validator v = new Validator();
        v.check(course.isPublished, "course", "course is not published");
        v.check(course.price == 0, "course", "this course is not free, please purchase it via checkout");
        v.throwIfInvalid();

        EnrollmentRepository.InsertResult result = enrollments.insert(claims.id(), id);

        service.grantCourseAccess(claims.id(), id);

        if (result.isNew()) {
            service.markEnrolled(id);
            service.notify(claims.id(), "course", "Enrolled in a course",
                    "You are now enrolled in \"" + course.title + "\". Happy learning!");
        }

        return ResponseEntity.status(HttpStatus.CREATED)
                .body(Map.of("enrollment", result.enrollment()));
    }

    @JsonNaming(PropertyNamingStrategies.SnakeCaseStrategy.class)
    record UpdateProgressRequest(long lessonId, boolean completed) {
    }

    // PATCH /v1/enrollments/{id}/progress. Body: {"lesson_id": N, "completed": bool}.
    @PatchMapping("/v1/enrollments/{id}/progress")
    Map<String, Object> updateProgress(@PathVariable long id, @RequestBody UpdateProgressRequest input) {
        UserPrincipal claims = CurrentUser.get();

        EnrollmentDto enrollment = enrollments.findById(id).orElseThrow(NotFoundException::new);

        if (enrollment.userId != claims.id()) {
            throw new NotPermittedException();
        }

        Validator v = new Validator();
        v.check(input.lessonId() > 0, "lesson_id", "must be provided");
        v.throwIfInvalid();

        try {
            enrollments.setLessonCompleted(claims.id(), input.lessonId(), input.completed());
        } catch (NotFoundException e) {
            v.addError("lesson_id", "you don't have access to this lesson");
            v.throwIfInvalid();
        }

        if (input.completed()) {
            service.maybeIssueCertificate(claims.id(), enrollment.courseId);
        }

        return Map.of("message", "progress updated");
    }
}
