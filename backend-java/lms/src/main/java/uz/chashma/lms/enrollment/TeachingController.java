package uz.chashma.lms.enrollment;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;
import uz.chashma.lms.course.api.CourseApi;
import uz.chashma.lms.course.api.CourseApi.LessonInfo;
import uz.chashma.lms.course.api.CourseDto;
import uz.chashma.lms.shared.security.CurrentUser;
import uz.chashma.lms.shared.security.UserPrincipal;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Go enrollment-service cmd/api/teaching.go porti — studio
 * dashboard/analytics ko'rsatkichlari.
 */
@RestController
class TeachingController {

    private static final Logger log = LoggerFactory.getLogger(TeachingController.class);

    private final EnrollmentRepository enrollments;
    private final OrderRepository orders;
    private final CourseApi courseApi;

    TeachingController(EnrollmentRepository enrollments, OrderRepository orders, CourseApi courseApi) {
        this.enrollments = enrollments;
        this.orders = orders;
        this.courseApi = courseApi;
    }

    /** Studio analytics "completion by course" qatori. */
    record CourseEngagement(long courseId, String title, int students, int completion) {
    }

    // GET /v1/me/teaching/stats.
    @GetMapping("/v1/me/teaching/stats")
    Map<String, Object> teachingStats() {
        UserPrincipal claims = CurrentUser.get();

        List<CourseDto> courses = courseApi.coursesByInstructor(claims.id());

        List<Long> courseIds = new ArrayList<>();
        int published = 0;
        int drafts = 0;
        double ratingSum = 0;
        int ratingCount = 0;
        for (CourseDto c : courses) {
            courseIds.add(c.id);
            if (c.isPublished) {
                published++;
            } else {
                drafts++;
            }
            if (c.ratingCount > 0) {
                ratingSum += c.rating;
                ratingCount++;
            }
        }

        // Alohida sotiladigan darslar ham daromadga kiradi.
        List<Long> lessonIds = new ArrayList<>();
        try {
            for (LessonInfo l : courseApi.lessonsForCourses(courseIds)) {
                lessonIds.add(l.id());
            }
        } catch (RuntimeException e) {
            log.warn("teaching stats: failed to fetch lessons: {}", e.getMessage());
        }

        OrderRepository.RevenueResult revenue = orders.revenueForItems(courseIds, lessonIds);

        Map<Long, Integer> studentCounts = enrollments.countsByCourses(courseIds);
        int totalStudents = enrollments.distinctStudentsForCourses(courseIds);
        EnrollmentRepository.CompletedStats completedStats = enrollments.completedStatsByCourses(courseIds);

        double avgQuizScore = 0;
        try {
            avgQuizScore = courseApi.avgQuizScore(courseIds);
        } catch (RuntimeException e) {
            log.warn("teaching stats: failed to fetch quiz stats: {}", e.getMessage());
        }

        // Har kurs bo'yicha tugatilganlik: tugatilgan dars yozuvlari /
        // (studentlar * darslar soni).
        List<CourseEngagement> engagement = new ArrayList<>();
        int completionSum = 0;
        int completionCourses = 0;
        for (CourseDto c : courses) {
            int students = studentCounts.getOrDefault(c.id, 0);
            int completion = 0;
            if (students > 0 && c.totalLessons > 0) {
                completion = completedStats.counts().getOrDefault(c.id, 0) * 100 / (students * c.totalLessons);
            }
            if (students > 0) {
                completionSum += completion;
                completionCourses++;
            }
            engagement.add(new CourseEngagement(c.id, c.title, students, completion));
        }

        int avgCompletion = completionCourses > 0 ? completionSum / completionCourses : 0;
        double avgRating = ratingCount > 0 ? ratingSum / ratingCount : 0;

        Map<String, Object> stats = new LinkedHashMap<>();
        stats.put("totalRevenue", revenue.total());
        stats.put("monthlyRevenue", revenue.monthly());
        stats.put("totalStudents", totalStudents);
        stats.put("activeStudents", completedStats.activeStudents());
        stats.put("publishedCourses", published);
        stats.put("draftCourses", drafts);
        stats.put("avgRating", avgRating);
        stats.put("avgCompletion", avgCompletion);
        stats.put("avgQuizScore", avgQuizScore);
        stats.put("engagement", engagement);
        return stats;
    }
}
