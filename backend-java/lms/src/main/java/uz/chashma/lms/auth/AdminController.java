package uz.chashma.lms.auth;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PatchMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;
import uz.chashma.lms.course.api.CourseApi;
import uz.chashma.lms.enrollment.api.EnrollmentApi;
import uz.chashma.lms.shared.BadRequestException;
import uz.chashma.lms.shared.UiDefaults;
import uz.chashma.lms.shared.Validator;
import uz.chashma.lms.shared.security.CurrentUser;
import uz.chashma.lms.shared.security.Roles;

import java.time.ZoneOffset;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/** Go auth-service cmd/api/admin.go porti (RBAC — SecurityConfig'da). */
@RestController
@RequestMapping("/v1/admin")
class AdminController {

    private static final DateTimeFormatter DATE = DateTimeFormatter.ofPattern("yyyy-MM-dd")
            .withZone(ZoneOffset.UTC);

    private final UserRepository users;
    private final CourseApi courseApi;
    private final EnrollmentApi enrollmentApi;

    AdminController(UserRepository users, CourseApi courseApi, EnrollmentApi enrollmentApi) {
        this.users = users;
        this.courseApi = courseApi;
        this.enrollmentApi = enrollmentApi;
    }

    // Frontend AdminUser shakli (services/mock/users.ts).
    record AdminUser(long id, String name, String email, String role, String avatarColor,
                     String joinedAt, String status, int coursesCreated, int coursesEnrolled) {
    }

    // GET /v1/admin/users?page=&pageSize=
    @GetMapping("/users")
    Map<String, Object> listUsers(@RequestParam(required = false) String page,
                                  @RequestParam(required = false) String pageSize) {
        int pageN = parseOr(page, 1);
        int pageSizeN = parseOr(pageSize, 20);

        Validator v = new Validator();
        v.check(pageN > 0, "page", "must be greater than zero");
        v.check(pageSizeN > 0 && pageSizeN <= 100, "pageSize", "must be between 1 and 100");
        v.throwIfInvalid();

        UserRepository.UserPage result = users.list(pageN, pageSizeN);

        List<Long> ids = result.users().stream().map(u -> u.id).toList();
        Map<Long, Integer> createdCounts = ids.isEmpty() ? Map.of() : courseApi.courseCountsByInstructor(ids);
        Map<Long, Integer> enrolledCounts = ids.isEmpty() ? Map.of() : enrollmentApi.enrollmentCountsByUser(ids);

        List<AdminUser> items = new ArrayList<>();
        for (User u : result.users()) {
            items.add(new AdminUser(
                    u.id, u.name, u.email, u.role,
                    UiDefaults.avatarColor(u.id),
                    DATE.format(u.createdAt),
                    "active",
                    createdCounts.getOrDefault(u.id, 0),
                    enrolledCounts.getOrDefault(u.id, 0)));
        }

        Map<String, Object> response = new LinkedHashMap<>();
        response.put("items", items);
        response.put("page", pageN);
        response.put("pageSize", pageSizeN);
        response.put("total", result.total());
        return response;
    }

    record UpdateRoleRequest(String role) {
    }

    // PATCH /v1/admin/users/{id}/role — admin/instructor/student tayinlash.
    @PatchMapping("/users/{id}/role")
    Map<String, Object> updateRole(@PathVariable long id, @RequestBody UpdateRoleRequest input) {
        String role = input.role() == null ? "" : input.role().trim();

        Validator v = new Validator();
        v.check(Validator.permitted(role, Roles.STUDENT, Roles.INSTRUCTOR, Roles.ADMIN),
                "role", "must be one of student, instructor, admin");
        v.throwIfInvalid();

        // Admin o'zini o'zi rolidan mahrum qilib qulflanib qolmasin.
        if (CurrentUser.get().id() == id) {
            throw new BadRequestException("you cannot change your own role");
        }

        users.updateRole(id, role);
        return Map.of("message", "user role updated to " + role);
    }

    // GET /v1/admin/stats (frontend AdminStats shakli).
    @GetMapping("/stats")
    Map<String, Object> stats() {
        CourseApi.CourseStats courseStats = courseApi.stats();

        Map<String, Object> stats = new LinkedHashMap<>();
        stats.put("totalUsers", users.count());
        stats.put("totalCourses", courseStats.totalCourses());
        stats.put("revenue", enrollmentApi.revenue());
        stats.put("activeInstructors", courseStats.activeInstructors());
        return stats;
    }

    // Go jsonutil.ReadInt kabi: yaroqsiz qiymat — default.
    private static int parseOr(String s, int defaultValue) {
        if (s == null || s.isBlank()) {
            return defaultValue;
        }
        try {
            return Integer.parseInt(s);
        } catch (NumberFormatException e) {
            return defaultValue;
        }
    }
}
