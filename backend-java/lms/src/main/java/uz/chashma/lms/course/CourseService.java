package uz.chashma.lms.course;

import org.springframework.stereotype.Service;
import uz.chashma.lms.auth.api.UserApi;
import uz.chashma.lms.course.api.CourseApi;
import uz.chashma.lms.course.api.CourseDto;
import uz.chashma.lms.course.api.InstructorDto;
import uz.chashma.lms.shared.UiDefaults;

import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;

/**
 * CourseApi implementatsiyasi + kurslarni dekoratsiyalash (Go
 * decorateCourses ekvivalenti: instruktor obyekti va UI-default maydonlar).
 */
@Service
class CourseService implements CourseApi {

    private final CourseRepository courses;
    private final QuizRepository quizzes;
    private final UserApi userApi;

    CourseService(CourseRepository courses, QuizRepository quizzes, UserApi userApi) {
        this.courses = courses;
        this.quizzes = quizzes;
        this.userApi = userApi;
    }

    /** Instruktor obyekti va UI-default maydonlarni to'ldiradi. */
    void decorate(List<CourseDto> list) {
        Set<Long> idSet = new HashSet<>();
        for (CourseDto c : list) {
            idSet.add(c.instructorId);
        }

        Map<Long, UserApi.UserSummary> users = new HashMap<>();
        for (UserApi.UserSummary u : userApi.findByIds(List.copyOf(idSet))) {
            users.put(u.id(), u);
        }

        for (CourseDto c : list) {
            c.thumbnailColor = UiDefaults.thumbnailColor(c.id);

            UserApi.UserSummary user = users.get(c.instructorId);

            InstructorDto instructor = new InstructorDto();
            instructor.id = c.instructorId;
            instructor.name = user != null ? user.name() : "Instructor";
            instructor.headline = "Instructor";
            instructor.avatarColor = UiDefaults.avatarColor(c.instructorId);
            c.instructor = instructor;
        }
    }

    CourseRepository.ListResult list(CourseRepository.CourseFilters filters) {
        CourseRepository.ListResult result = courses.list(filters);
        decorate(result.courses());
        return result;
    }

    @Override
    public List<CourseDto> coursesByIds(List<Long> ids) {
        if (ids == null || ids.isEmpty()) {
            return List.of();
        }
        CourseRepository.ListResult result = courses.list(new CourseRepository.CourseFilters(
                "", "", "", 1, ids.size(), ids, 0, true));
        decorate(result.courses());
        return result.courses();
    }

    @Override
    public List<CourseDto> coursesByInstructor(long instructorId) {
        CourseRepository.ListResult result = courses.list(new CourseRepository.CourseFilters(
                "", "", "", 1, 1000, List.of(), instructorId, true));
        decorate(result.courses());
        return result.courses();
    }

    @Override
    public List<LessonInfo> lessonsForCourse(long courseId) {
        return courses.lessonsForCourses(List.of(courseId));
    }

    @Override
    public List<LessonInfo> lessonsForCourses(List<Long> courseIds) {
        return courses.lessonsForCourses(courseIds);
    }

    @Override
    public List<LessonInfo> lessonsByIds(List<Long> ids) {
        return courses.lessonsByIds(ids);
    }

    @Override
    public void incrementStudentCount(long courseId) {
        courses.incrementStudentCount(courseId);
    }

    @Override
    public double avgQuizScore(List<Long> courseIds) {
        return quizzes.avgScoreForCourses(courseIds);
    }

    @Override
    public CourseStats stats() {
        return courses.stats();
    }

    @Override
    public Map<Long, Integer> courseCountsByInstructor(List<Long> userIds) {
        return courses.courseCountsByInstructor(userIds);
    }
}
