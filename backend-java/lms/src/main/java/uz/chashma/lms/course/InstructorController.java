package uz.chashma.lms.course;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RestController;
import uz.chashma.lms.auth.api.UserApi;
import uz.chashma.lms.course.api.InstructorDto;
import uz.chashma.lms.shared.NotFoundException;
import uz.chashma.lms.shared.UiDefaults;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * Go course-service cmd/api/instructors.go porti. Instruktorlar alohida
 * jadval emas — kamida bitta published kursi bor userlar. Statistika
 * courses jadvalidan, ismlar auth modulidan.
 */
@RestController
class InstructorController {

    private final CourseRepository courses;
    private final UserApi userApi;

    InstructorController(CourseRepository courses, UserApi userApi) {
        this.courses = courses;
        this.userApi = userApi;
    }

    @GetMapping("/v1/instructors")
    Map<String, Object> list() {
        List<CourseRepository.InstructorStat> stats = courses.instructorStats();

        List<Long> ids = stats.stream().map(CourseRepository.InstructorStat::instructorId).toList();
        Map<Long, UserApi.UserSummary> users = fetchUsers(ids);

        List<InstructorDto> items = new ArrayList<>();
        for (CourseRepository.InstructorStat s : stats) {
            UserApi.UserSummary user = users.get(s.instructorId());
            InstructorDto instructor = new InstructorDto();
            instructor.id = s.instructorId();
            instructor.name = user != null ? user.name() : "Instructor";
            instructor.headline = "Instructor";
            instructor.avatarColor = UiDefaults.avatarColor(s.instructorId());
            instructor.students = s.students();
            instructor.courses = s.courseCount();
            instructor.rating = s.rating();
            items.add(instructor);
        }

        return Map.of("items", items);
    }

    @GetMapping("/v1/instructors/{id}")
    Map<String, Object> show(@PathVariable long id) {
        List<CourseRepository.InstructorStat> stats = courses.instructorStats();

        CourseRepository.InstructorStat stat = new CourseRepository.InstructorStat(id, 0, 0, 0);
        for (CourseRepository.InstructorStat s : stats) {
            if (s.instructorId() == id) {
                stat = s;
                break;
            }
        }

        Map<Long, UserApi.UserSummary> users = fetchUsers(List.of(id));
        UserApi.UserSummary user = users.get(id);
        if (user == null) {
            throw new NotFoundException();
        }

        InstructorDto instructor = new InstructorDto();
        instructor.id = id;
        instructor.name = user.name();
        instructor.headline = "Instructor";
        instructor.avatarColor = UiDefaults.avatarColor(id);
        instructor.students = stat.students();
        instructor.courses = stat.courseCount();
        instructor.rating = stat.rating();

        return Map.of("instructor", instructor);
    }

    private Map<Long, UserApi.UserSummary> fetchUsers(List<Long> ids) {
        Map<Long, UserApi.UserSummary> users = new HashMap<>();
        for (UserApi.UserSummary u : userApi.findByIds(ids)) {
            users.put(u.id(), u);
        }
        return users;
    }
}
