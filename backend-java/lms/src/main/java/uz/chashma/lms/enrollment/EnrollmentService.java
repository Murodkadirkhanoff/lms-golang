package uz.chashma.lms.enrollment;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import uz.chashma.lms.course.api.CourseApi;
import uz.chashma.lms.course.api.CourseDto;

import java.util.List;

/**
 * Modul ichidagi umumiy oqimlar (Go handlerlaridagi yordamchi metodlar):
 * kirish berish, bildirishnoma, sertifikat.
 */
@Service
class EnrollmentService {

    private static final Logger log = LoggerFactory.getLogger(EnrollmentService.class);

    private final EnrollmentRepository enrollments;
    private final CertificateRepository certificates;
    private final NotificationRepository notifications;
    private final CourseApi courseApi;

    EnrollmentService(EnrollmentRepository enrollments, CertificateRepository certificates,
                      NotificationRepository notifications, CourseApi courseApi) {
        this.enrollments = enrollments;
        this.certificates = certificates;
        this.notifications = notifications;
        this.courseApi = courseApi;
    }

    /**
     * Kursning barcha darslariga kirish beradi (Go grantCourseAccess —
     * eski sync_lesson_access_on_enrollment triggerining o'rni).
     */
    void grantCourseAccess(long userId, long courseId) {
        List<Long> lessonIds = courseApi.lessonsForCourse(courseId).stream()
                .map(CourseApi.LessonInfo::id)
                .toList();
        enrollments.grantLessonAccess(userId, courseId, lessonIds);
    }

    /** student_count'ni oshiradi; xato asosiy oqimni yiqitmaydi. */
    void markEnrolled(long courseId) {
        try {
            courseApi.incrementStudentCount(courseId);
        } catch (RuntimeException e) {
            log.warn("failed to increment student count: courseId={} error={}", courseId, e.getMessage());
        }
    }

    /** Xabarnoma yozadi; xato bo'lsa so'rovni yiqitmaydi, faqat log. */
    void notify(long userId, String type, String title, String body) {
        try {
            notifications.insert(userId, type, title, body);
        } catch (RuntimeException e) {
            log.warn("failed to insert notification: {}", e.getMessage());
        }
    }

    /**
     * Kursning barcha darslari tugatilgan bo'lsa sertifikat beradi.
     * Yordamchi oqim — xatolar progress so'rovini yiqitmaydi.
     */
    void maybeIssueCertificate(long userId, long courseId) {
        CourseDto course;
        try {
            List<CourseDto> courses = courseApi.coursesByIds(List.of(courseId));
            if (courses.isEmpty()) {
                return;
            }
            course = courses.get(0);
        } catch (RuntimeException e) {
            log.warn("certificate check: failed to fetch course: {}", e.getMessage());
            return;
        }

        if (course.totalLessons == 0) {
            return;
        }

        int completed;
        try {
            completed = enrollments.completedCounts(userId).getOrDefault(courseId, 0);
        } catch (RuntimeException e) {
            log.warn("certificate check: failed to count lessons: {}", e.getMessage());
            return;
        }

        if (completed < course.totalLessons) {
            return;
        }

        boolean issued;
        try {
            issued = certificates.issue(userId, courseId, course.title);
        } catch (RuntimeException e) {
            log.warn("failed to issue certificate: {}", e.getMessage());
            return;
        }

        if (issued) {
            notify(userId, "course", "Certificate earned",
                    "Congratulations! You completed \"" + course.title + "\" and earned a certificate.");
        }
    }
}
