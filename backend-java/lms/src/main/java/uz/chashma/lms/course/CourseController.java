package uz.chashma.lms.course;

import tools.jackson.databind.PropertyNamingStrategies;
import tools.jackson.databind.annotation.JsonNaming;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PatchMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;
import uz.chashma.lms.course.api.CourseDto;
import uz.chashma.lms.course.api.LessonDto;
import uz.chashma.lms.course.api.ModuleDto;
import uz.chashma.lms.enrollment.api.EnrollmentApi;
import uz.chashma.lms.shared.Ids;
import uz.chashma.lms.shared.NotFoundException;
import uz.chashma.lms.shared.NotPermittedException;
import uz.chashma.lms.shared.Validator;
import uz.chashma.lms.shared.security.CurrentUser;
import uz.chashma.lms.shared.security.Roles;
import uz.chashma.lms.shared.security.UserPrincipal;

import java.net.URI;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;

import static uz.chashma.lms.shared.Validator.byteLength;
import static uz.chashma.lms.shared.Validator.orEmpty;

/** Go course-service cmd/api/courses.go porti. */
@RestController
class CourseController {

    private static final Logger log = LoggerFactory.getLogger(CourseController.class);

    private final CourseService service;
    private final CourseRepository courses;
    private final ReviewRepository reviews;
    private final EnrollmentApi enrollmentApi;

    CourseController(CourseService service, CourseRepository courses,
                     ReviewRepository reviews, EnrollmentApi enrollmentApi) {
        this.service = service;
        this.courses = courses;
        this.reviews = reviews;
        this.enrollmentApi = enrollmentApi;
    }

    // GET /v1/courses
    @GetMapping("/v1/courses")
    Map<String, Object> list(@RequestParam(defaultValue = "") String search,
                             @RequestParam(defaultValue = "") String category,
                             @RequestParam(defaultValue = "popular") String sort,
                             @RequestParam(defaultValue = "") String page,
                             @RequestParam(defaultValue = "") String pageSize,
                             @RequestParam(defaultValue = "") String ids,
                             @RequestParam(defaultValue = "") String instructorId) {
        int pageN = parseOr(page, 1);
        int pageSizeN = parseOr(pageSize, 8);
        long instructorIdN = parseOr(instructorId, 0);
        List<Long> idList = Ids.parse(ids);

        Validator v = new Validator();
        v.check(pageN > 0, "page", "must be greater than zero");
        v.check(pageSizeN > 0 && pageSizeN <= 100, "pageSize", "must be between 1 and 100");
        v.check(Validator.permitted(sort, "popular", "newest", "price-asc", "price-desc"),
                "sort", "must be one of popular, newest, price-asc, price-desc");
        v.throwIfInvalid();

        // Studio o'z kurslarini (draft ham) instructorId bilan so'raydi.
        boolean includeUnpublished = instructorIdN != 0;

        CourseRepository.ListResult result = service.list(new CourseRepository.CourseFilters(
                search, category, sort, pageN, pageSizeN, idList, instructorIdN, includeUnpublished));

        Map<String, Object> response = new LinkedHashMap<>();
        response.put("items", result.courses());
        response.put("page", pageN);
        response.put("pageSize", pageSizeN);
        response.put("total", result.total());
        return response;
    }

    // GET /v1/courses/{idOrSlug}
    @GetMapping("/v1/courses/{idOrSlug}")
    Map<String, Object> show(@PathVariable String idOrSlug) {
        CourseDto course = courses.findByIdOrSlug(idOrSlug).orElseThrow(NotFoundException::new);

        course.reviews = reviews.listForCourse(course.id, 20);

        service.decorate(List.of(course));
        sanitizeCourseContent(course);

        return Map.of("course", course);
    }

    @JsonNaming(PropertyNamingStrategies.SnakeCaseStrategy.class)
    record LessonRequest(String title, String type, String contentUrl, String content,
                         int durationSeconds, int position, double price, boolean isFree) {
    }

    @JsonNaming(PropertyNamingStrategies.SnakeCaseStrategy.class)
    record ModuleRequest(String title, int position, List<LessonRequest> lessons) {
    }

    @JsonNaming(PropertyNamingStrategies.SnakeCaseStrategy.class)
    record CreateCourseRequest(String title, String description, String thumbnailUrl, Long categoryId,
                               String lang, double price, boolean isPublished, List<ModuleRequest> modules) {
    }

    // POST /v1/courses. Request kalitlari snake_case, instructor_id tokendan.
    @PostMapping("/v1/courses")
    ResponseEntity<Map<String, Object>> create(@RequestBody CreateCourseRequest input) {
        UserPrincipal claims = CurrentUser.get();

        CourseDto course = new CourseDto();
        course.title = orEmpty(input.title());
        course.description = orEmpty(input.description());
        course.thumbnailUrl = orEmpty(input.thumbnailUrl());
        course.categoryId = input.categoryId();
        course.lang = orEmpty(input.lang());
        course.price = input.price();
        course.isPublished = input.isPublished();
        course.instructorId = claims.id();
        course.modules = toModules(input.modules());

        Validator v = new Validator();
        validateCourse(v, course);
        v.throwIfInvalid();

        course.slug = Slugs.slugify(course.title);
        if (course.slug.isEmpty()) {
            v.addError("title", "must contain at least one latin letter or digit for slug generation");
            v.throwIfInvalid();
        }

        try {
            courses.insert(course);
        } catch (CourseErrors.InvalidParentException e) {
            v.addError("category_id", "category does not exist");
            v.throwIfInvalid();
        }

        service.decorate(List.of(course));

        return ResponseEntity.created(URI.create("/v1/courses/" + course.id))
                .body(Map.of("course", course));
    }

    @JsonNaming(PropertyNamingStrategies.SnakeCaseStrategy.class)
    record UpdateCourseRequest(String title, String description, String thumbnailUrl, Long categoryId,
                               String lang, Double price, Boolean isPublished,
                               // null — o'quv rejasiga tegilmaydi; berilsa butunlay almashtiriladi.
                               List<ModuleRequest> modules) {
    }

    // PATCH /v1/courses/{id}
    @PatchMapping("/v1/courses/{id}")
    Map<String, Object> update(@PathVariable long id, @RequestBody UpdateCourseRequest input) {
        CourseDto course = courses.findByIdOrSlug(Long.toString(id)).orElseThrow(NotFoundException::new);

        if (!canModifyCourse(course)) {
            throw new NotPermittedException();
        }

        if (input.title() != null) {
            course.title = input.title();
        }
        if (input.description() != null) {
            course.description = input.description();
        }
        if (input.thumbnailUrl() != null) {
            course.thumbnailUrl = input.thumbnailUrl();
        }
        if (input.categoryId() != null) {
            course.categoryId = input.categoryId();
        }
        if (input.lang() != null) {
            course.lang = input.lang();
        }
        if (input.price() != null) {
            course.price = input.price();
        }
        if (input.isPublished() != null) {
            course.isPublished = input.isPublished();
        }

        List<ModuleDto> newModules = null;
        if (input.modules() != null) {
            newModules = toModules(input.modules());
            course.modules = newModules;
        }

        Validator v = new Validator();
        validateCourse(v, course);
        v.throwIfInvalid();

        try {
            courses.update(course);
        } catch (CourseErrors.InvalidParentException e) {
            v.addError("category_id", "category does not exist");
            v.throwIfInvalid();
        }

        if (newModules != null) {
            courses.replaceModules(course.id, newModules);
        }

        // To'liq yangilangan holatni qaytaramiz (modules + aggregatlar bilan).
        CourseDto updated = courses.findByIdOrSlug(Long.toString(course.id))
                .orElseThrow(NotFoundException::new);
        service.decorate(List.of(updated));

        return Map.of("course", updated);
    }

    // DELETE /v1/courses/{id}
    @DeleteMapping("/v1/courses/{id}")
    Map<String, Object> delete(@PathVariable long id) {
        CourseDto course = courses.findByIdOrSlug(Long.toString(id)).orElseThrow(NotFoundException::new);

        if (!canModifyCourse(course)) {
            throw new NotPermittedException();
        }

        courses.delete(id);

        return Map.of("message", "course successfully deleted");
    }

    /**
     * Paywall (Go sanitizeCourseContent): bepul bo'lmagan darslarning
     * kontentini kirish huquqi bo'lmagan so'rovchidan yashiradi. Kurs egasi
     * va admin to'liq ko'radi.
     */
    private void sanitizeCourseContent(CourseDto course) {
        if (course.modules == null || course.modules.isEmpty() || canModifyCourse(course)) {
            return;
        }

        Set<Long> accessible = new HashSet<>();
        UserPrincipal claims = CurrentUser.get();
        if (claims != null) {
            try {
                accessible.addAll(enrollmentApi.accessibleLessonIds(claims.id(), course.id));
            } catch (RuntimeException e) {
                // Fail closed: faqat bepul darslar ochiq qoladi.
                log.warn("paywall: failed to fetch lesson access: {}", e.getMessage());
            }
        }

        for (ModuleDto module : course.modules) {
            for (LessonDto lesson : module.lessons) {
                if (lesson.isFree || accessible.contains(lesson.id)) {
                    continue;
                }
                lesson.content = "";
                lesson.contentUrl = "";
                lesson.locked = true;
            }
        }
    }

    /** Faqat kurs egasi yoki admin o'zgartira oladi. */
    private static boolean canModifyCourse(CourseDto course) {
        UserPrincipal claims = CurrentUser.get();
        if (claims == null) {
            return false;
        }
        return claims.id() == course.instructorId || Roles.ADMIN.equals(claims.role());
    }

    private static List<ModuleDto> toModules(List<ModuleRequest> input) {
        List<ModuleDto> modules = new ArrayList<>();
        if (input == null) {
            return modules;
        }
        for (ModuleRequest m : input) {
            ModuleDto module = new ModuleDto();
            module.title = orEmpty(m.title());
            module.position = m.position();
            if (m.lessons() != null) {
                for (LessonRequest l : m.lessons()) {
                    LessonDto lesson = new LessonDto();
                    lesson.title = orEmpty(l.title());
                    lesson.type = (l.type() == null || l.type().isEmpty()) ? "video" : l.type();
                    lesson.contentUrl = orEmpty(l.contentUrl());
                    lesson.content = orEmpty(l.content());
                    lesson.durationSeconds = l.durationSeconds();
                    lesson.position = l.position();
                    lesson.price = l.price();
                    lesson.isFree = l.isFree();
                    module.lessons.add(lesson);
                }
            }
            modules.add(module);
        }
        return modules;
    }

    // Go data.ValidateCourse bilan bir xil xabarlar.
    static void validateCourse(Validator v, CourseDto course) {
        v.check(!course.title.isEmpty(), "title", "must be provided");
        v.check(byteLength(course.title) <= 200, "title", "must not be more than 200 bytes long");
        v.check(Validator.permitted(course.lang, "uz", "ru", "en"), "lang", "must be one of uz, ru, en");
        v.check(course.price >= 0, "price", "must not be negative");

        if (course.modules == null) {
            return;
        }
        for (int mi = 0; mi < course.modules.size(); mi++) {
            ModuleDto module = course.modules.get(mi);
            v.check(!module.title.isEmpty(), "modules[%d].title".formatted(mi), "must be provided");
            for (int li = 0; li < module.lessons.size(); li++) {
                LessonDto lesson = module.lessons.get(li);
                String key = "modules[%d].lessons[%d]".formatted(mi, li);
                v.check(!lesson.title.isEmpty(), key + ".title", "must be provided");
                v.check(Validator.permitted(lesson.type, "video", "text"), key + ".type", "must be one of video, text");
                v.check(lesson.price >= 0, key + ".price", "must not be negative");
                v.check(!lesson.isFree || lesson.price == 0, key + ".price", "free lessons must have price 0");
                v.check(lesson.durationSeconds >= 0, key + ".durationSeconds", "must not be negative");
            }
        }
    }

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
