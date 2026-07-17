package uz.chashma.lms.course;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;
import uz.chashma.lms.auth.api.UserApi;
import uz.chashma.lms.course.api.ReviewDto;
import uz.chashma.lms.enrollment.api.EnrollmentApi;
import uz.chashma.lms.shared.NotFoundException;
import uz.chashma.lms.shared.NotPermittedException;
import uz.chashma.lms.shared.Validator;
import uz.chashma.lms.shared.security.CurrentUser;
import uz.chashma.lms.shared.security.UserPrincipal;

import java.util.List;
import java.util.Map;

import static uz.chashma.lms.shared.Validator.byteLength;
import static uz.chashma.lms.shared.Validator.orEmpty;

/**
 * Go course-service cmd/api/reviews.go porti. Bitta user bitta kursga bitta
 * sharh; qayta yuborsa yangilanadi.
 */
@RestController
class ReviewController {

    private final ReviewRepository reviews;
    private final UserApi userApi;
    private final EnrollmentApi enrollmentApi;

    ReviewController(ReviewRepository reviews, UserApi userApi, EnrollmentApi enrollmentApi) {
        this.reviews = reviews;
        this.userApi = userApi;
        this.enrollmentApi = enrollmentApi;
    }

    record CreateReviewRequest(int rating, String comment) {
    }

    // POST /v1/courses/{id}/reviews (auth).
    @PostMapping("/v1/courses/{id}/reviews")
    ResponseEntity<Map<String, Object>> create(@PathVariable long id,
                                               @RequestBody CreateReviewRequest input) {
        UserPrincipal claims = CurrentUser.get();

        // Faqat kursga yozilgan foydalanuvchi baho qo'ya oladi.
        if (!enrollmentApi.isEnrolled(claims.id(), id)) {
            throw new NotPermittedException("you must be enrolled in this course to leave a review");
        }

        // Ism snapshot uchun auth modulidan olinadi.
        String userName = "Student";
        List<UserApi.UserSummary> users = userApi.findByIds(List.of(claims.id()));
        if (!users.isEmpty()) {
            userName = users.get(0).name();
        }

        ReviewDto review = new ReviewDto();
        review.courseId = id;
        review.userId = claims.id();
        review.user = userName;
        review.rating = input.rating();
        review.comment = orEmpty(input.comment());

        Validator v = new Validator();
        v.check(review.rating >= 1 && review.rating <= 5, "rating", "must be between 1 and 5");
        v.check(byteLength(review.comment) <= 2000, "comment", "must not be more than 2000 bytes long");
        v.throwIfInvalid();

        try {
            reviews.upsert(review);
        } catch (CourseErrors.InvalidCourseException e) {
            throw new NotFoundException();
        }

        return ResponseEntity.status(HttpStatus.CREATED).body(Map.of("review", review));
    }
}
