package uz.chashma.lms.course.api;

import java.util.List;
import java.util.Map;

/**
 * course modulining public facade'i (Go'dagi /internal/* endpointlar o'rni).
 * Microservice'ga ajratilganda REST/gRPC client bilan almashadi.
 */
public interface CourseApi {

    /** To'liq (instruktor bilan dekoratsiyalangan) kurslar, id bo'yicha. */
    List<CourseDto> coursesByIds(List<Long> ids);

    /** Instruktorning barcha kurslari (unpublished ham — studio uchun). */
    List<CourseDto> coursesByInstructor(long instructorId);

    /** Kursning darslari modul/dars tartibida. */
    List<LessonInfo> lessonsForCourse(long courseId);

    /** Bir nechta kursning darslari tartib bilan (currentLesson uchun). */
    List<LessonInfo> lessonsForCourses(List<Long> courseIds);

    /** Alohida sotib olinayotgan darslar (checkout). */
    List<LessonInfo> lessonsByIds(List<Long> ids);

    /** Yangi enrollment'da student_count'ni oshiradi. */
    void incrementStudentCount(long courseId);

    /** Kurslar bo'yicha o'rtacha quiz bali (instruktor analitikasi). */
    double avgQuizScore(List<Long> courseIds);

    /** Admin panel: published kurslar va faol instruktorlar soni. */
    CourseStats stats();

    /** Admin users ro'yxati: user -> yaratgan kurslari soni. */
    Map<Long, Integer> courseCountsByInstructor(List<Long> userIds);

    record CourseStats(int totalCourses, int activeInstructors) {
    }

    record LessonInfo(long id, String title, double price, boolean isFree,
                      long courseId, String courseTitle) {
    }
}
