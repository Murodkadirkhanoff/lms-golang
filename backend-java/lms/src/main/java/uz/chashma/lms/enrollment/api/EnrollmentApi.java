package uz.chashma.lms.enrollment.api;

import java.util.List;
import java.util.Map;

/**
 * enrollment modulining public facade'i (Go'dagi /internal/* endpointlar o'rni).
 */
public interface EnrollmentApi {

    /** Barcha to'langan buyurtmalar summasi (admin stats). */
    double revenue();

    /** Admin users ro'yxati: user -> yozilgan kurslari soni. */
    Map<Long, Integer> enrollmentCountsByUser(List<Long> userIds);

    /** Paywall: foydalanuvchining kursda kirish huquqi bor dars id'lari. */
    List<Long> accessibleLessonIds(long userId, long courseId);

    /** Review sharti: foydalanuvchi kursga yozilganmi. */
    boolean isEnrolled(long userId, long courseId);
}
