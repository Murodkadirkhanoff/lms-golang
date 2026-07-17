package uz.chashma.lms.enrollment;

import org.springframework.stereotype.Service;
import uz.chashma.lms.enrollment.api.EnrollmentApi;

import java.util.List;
import java.util.Map;

@Service
class EnrollmentApiImpl implements EnrollmentApi {

    private final EnrollmentRepository enrollments;
    private final OrderRepository orders;

    EnrollmentApiImpl(EnrollmentRepository enrollments, OrderRepository orders) {
        this.enrollments = enrollments;
        this.orders = orders;
    }

    @Override
    public double revenue() {
        return orders.revenue();
    }

    @Override
    public Map<Long, Integer> enrollmentCountsByUser(List<Long> userIds) {
        return enrollments.enrollmentCountsByUser(userIds);
    }

    @Override
    public List<Long> accessibleLessonIds(long userId, long courseId) {
        return enrollments.accessibleLessonIds(userId, courseId);
    }

    @Override
    public boolean isEnrolled(long userId, long courseId) {
        return !enrollments.ownedCourses(userId, List.of(courseId)).isEmpty();
    }
}
