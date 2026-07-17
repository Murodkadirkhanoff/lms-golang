package uz.chashma.lms.enrollment;

import tools.jackson.databind.PropertyNamingStrategies;
import tools.jackson.databind.annotation.JsonNaming;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;
import uz.chashma.lms.auth.api.UserApi;
import uz.chashma.lms.course.api.CourseApi;
import uz.chashma.lms.course.api.CourseApi.LessonInfo;
import uz.chashma.lms.course.api.CourseDto;
import uz.chashma.lms.shared.NotFoundException;
import uz.chashma.lms.shared.UiDefaults;
import uz.chashma.lms.shared.Validator;
import uz.chashma.lms.shared.security.CurrentUser;
import uz.chashma.lms.shared.security.UserPrincipal;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;

/** Go enrollment-service me.go + enrollments.go (me qismlari) + orders.go porti. */
@RestController
@RequestMapping("/v1/me")
class MeController {

    private final EnrollmentRepository enrollments;
    private final OrderRepository orders;
    private final CertificateRepository certificates;
    private final NotificationRepository notifications;
    private final EnrollmentService service;
    private final CourseApi courseApi;
    private final UserApi userApi;

    MeController(EnrollmentRepository enrollments, OrderRepository orders,
                 CertificateRepository certificates, NotificationRepository notifications,
                 EnrollmentService service, CourseApi courseApi, UserApi userApi) {
        this.enrollments = enrollments;
        this.orders = orders;
        this.certificates = certificates;
        this.notifications = notifications;
        this.service = service;
        this.courseApi = courseApi;
        this.userApi = userApi;
    }

    // GET /v1/me/stats (dashboard DashboardStats shakli).
    @GetMapping("/stats")
    Map<String, Object> stats() {
        UserPrincipal claims = CurrentUser.get();

        List<EnrollmentDto> userEnrollments = enrollments.listByUser(claims.id());

        List<Long> courseIds = userEnrollments.stream().map(e -> e.courseId).toList();
        Map<Long, CourseDto> courses = coursesById(courseIds);
        Map<Long, Integer> completed = enrollments.completedCounts(claims.id());

        int completedCourses = 0;
        int inProgress = 0;
        for (EnrollmentDto e : userEnrollments) {
            CourseDto course = courses.get(e.courseId);
            if (course == null || course.totalLessons == 0) {
                continue;
            }
            int done = completed.getOrDefault(e.courseId, 0);
            if (done >= course.totalLessons) {
                completedCourses++;
            } else if (done > 0) {
                inProgress++;
            }
        }

        Map<String, Object> stats = new LinkedHashMap<>();
        stats.put("enrolled", userEnrollments.size());
        stats.put("inProgress", inProgress);
        stats.put("completed", completedCourses);
        stats.put("certificates", certificates.countByUser(claims.id()));
        return stats;
    }

    /** Frontend EnrolledCourse tipi: course — to'liq Course JSON'i. */
    record EnrolledCourse(long enrollmentId, CourseDto course, int progress, String currentLesson,
                          int lessonsCompleted, List<Long> completedLessonIds) {
    }

    // GET /v1/me/courses?page=&pageSize=
    @GetMapping("/courses")
    Map<String, Object> myCourses(@RequestParam(required = false) String page,
                                  @RequestParam(required = false) String pageSize) {
        UserPrincipal claims = CurrentUser.get();

        int pageN = parseOr(page, 1);
        int pageSizeN = parseOr(pageSize, 20);
        validatePage(pageN, pageSizeN);

        EnrollmentRepository.EnrollmentPage result = enrollments.pageByUser(claims.id(), pageN, pageSizeN);
        List<EnrollmentDto> userEnrollments = result.enrollments();

        List<Long> courseIds = userEnrollments.stream().map(e -> e.courseId).toList();
        Map<Long, CourseDto> courses = coursesById(courseIds);
        Map<Long, Integer> completed = enrollments.completedCounts(claims.id());

        // currentLesson: tartibdagi birinchi tugatilmagan dars nomi.
        Map<Long, List<LessonInfo>> lessonsByCourse = new HashMap<>();
        for (LessonInfo lesson : courseApi.lessonsForCourses(courseIds)) {
            lessonsByCourse.computeIfAbsent(lesson.courseId(), k -> new ArrayList<>()).add(lesson);
        }

        Set<Long> doneLessons = enrollments.completedLessonIds(claims.id());

        List<EnrolledCourse> items = new ArrayList<>();
        for (EnrollmentDto e : userEnrollments) {
            CourseDto course = courses.get(e.courseId);
            if (course == null) {
                continue; // kurs o'chirilgan bo'lishi mumkin
            }

            int done = completed.getOrDefault(e.courseId, 0);
            int progress = 0;
            if (course.totalLessons > 0) {
                progress = done * 100 / course.totalLessons;
            }

            String currentLesson = "";
            List<Long> completedIds = new ArrayList<>();
            for (LessonInfo lesson : lessonsByCourse.getOrDefault(e.courseId, List.of())) {
                if (doneLessons.contains(lesson.id())) {
                    completedIds.add(lesson.id());
                } else if (currentLesson.isEmpty()) {
                    currentLesson = lesson.title();
                }
            }

            items.add(new EnrolledCourse(e.id, course, progress, currentLesson, done, completedIds));
        }

        Map<String, Object> response = new LinkedHashMap<>();
        response.put("items", items);
        response.put("page", pageN);
        response.put("pageSize", pageSizeN);
        response.put("total", result.total());
        return response;
    }

    // GET /v1/me/certificates.
    @GetMapping("/certificates")
    Map<String, Object> myCertificates() {
        UserPrincipal claims = CurrentUser.get();
        return Map.of("items", certificates.listByUser(claims.id()));
    }

    // GET /v1/me/certificates/{id}/download — PDF fayl.
    @GetMapping("/certificates/{id}/download")
    ResponseEntity<byte[]> downloadCertificate(@PathVariable long id) {
        UserPrincipal claims = CurrentUser.get();

        CertificateRepository.CertificateDto cert = certificates.findForUser(id, claims.id())
                .orElseThrow(NotFoundException::new);

        String studentName = "Student";
        List<UserApi.UserSummary> found = userApi.findByIds(List.of(claims.id()));
        if (!found.isEmpty()) {
            studentName = found.get(0).name();
        }

        byte[] pdf = CertificatePdf.render(studentName, cert);

        return ResponseEntity.ok()
                .header(HttpHeaders.CONTENT_DISPOSITION,
                        "attachment; filename=\"certificate-LH-%06d.pdf\"".formatted(cert.id))
                .contentType(MediaType.APPLICATION_PDF)
                .body(pdf);
    }

    // GET /v1/me/notifications.
    @GetMapping("/notifications")
    Map<String, Object> myNotifications() {
        UserPrincipal claims = CurrentUser.get();
        return Map.of("items", notifications.listByUser(claims.id()));
    }

    // POST /v1/me/notifications/{id}/read — bittalab o'qildi qilish.
    @PostMapping("/notifications/{id}/read")
    Map<String, Object> readNotification(@PathVariable long id) {
        UserPrincipal claims = CurrentUser.get();
        notifications.markRead(id, claims.id());
        return Map.of("message", "notification marked as read");
    }

    // POST /v1/me/notifications/read-all.
    @PostMapping("/notifications/read-all")
    Map<String, Object> readAllNotifications() {
        UserPrincipal claims = CurrentUser.get();
        notifications.markAllRead(claims.id());
        return Map.of("message", "all notifications marked as read");
    }

    // GET /v1/me/orders?page=&pageSize=
    @GetMapping("/orders")
    Map<String, Object> myOrders(@RequestParam(required = false) String page,
                                 @RequestParam(required = false) String pageSize) {
        UserPrincipal claims = CurrentUser.get();

        int pageN = parseOr(page, 1);
        int pageSizeN = parseOr(pageSize, 20);
        validatePage(pageN, pageSizeN);

        OrderRepository.OrderPage result = orders.listByUser(claims.id(), pageN, pageSizeN);

        Map<String, Object> response = new LinkedHashMap<>();
        response.put("items", result.orders());
        response.put("page", pageN);
        response.put("pageSize", pageSizeN);
        response.put("total", result.total());
        return response;
    }

    // GET /v1/me/orders/{id}.
    @GetMapping("/orders/{id}")
    Map<String, Object> myOrder(@PathVariable long id) {
        UserPrincipal claims = CurrentUser.get();
        OrderDto order = orders.findForUser(id, claims.id()).orElseThrow(NotFoundException::new);
        return Map.of("order", order);
    }

    @JsonNaming(PropertyNamingStrategies.SnakeCaseStrategy.class)
    record CheckoutItem(Long courseId, Long lessonId) {
    }

    @JsonNaming(PropertyNamingStrategies.SnakeCaseStrategy.class)
    record CheckoutRequest(List<CheckoutItem> items, String paymentMethod) {
    }

    // POST /v1/me/orders. Narxlar clientdan emas, course modulidan olinadi.
    // To'lov hozircha mock: buyurtma darhol "paid" bo'ladi.
    @PostMapping("/orders")
    ResponseEntity<Map<String, Object>> checkout(@RequestBody CheckoutRequest input) {
        UserPrincipal claims = CurrentUser.get();

        List<CheckoutItem> inputItems = input.items() == null ? List.of() : input.items();
        String paymentMethod = (input.paymentMethod() == null || input.paymentMethod().isEmpty())
                ? "card" : input.paymentMethod();

        Validator v = new Validator();
        v.check(!inputItems.isEmpty(), "items", "must contain at least one item");

        // Takroriy elementlar savatda bo'lsa ham bir marta hisoblanadi.
        List<Long> courseIds = new ArrayList<>();
        List<Long> lessonIds = new ArrayList<>();
        Set<Long> seenCourses = new HashSet<>();
        Set<Long> seenLessons = new HashSet<>();
        for (int i = 0; i < inputItems.size(); i++) {
            CheckoutItem item = inputItems.get(i);
            boolean hasCourse = item.courseId() != null && item.courseId() > 0;
            boolean hasLesson = item.lessonId() != null && item.lessonId() > 0;
            if (hasCourse == hasLesson) {
                v.addError("items[%d]".formatted(i), "must contain exactly one of course_id or lesson_id");
                continue;
            }
            if (hasCourse && seenCourses.add(item.courseId())) {
                courseIds.add(item.courseId());
            }
            if (hasLesson && seenLessons.add(item.lessonId())) {
                lessonIds.add(item.lessonId());
            }
        }

        v.throwIfInvalid();

        Map<Long, CourseDto> courses = coursesById(courseIds);

        Map<Long, LessonInfo> lessons = new HashMap<>();
        for (LessonInfo lesson : courseApi.lessonsByIds(lessonIds)) {
            lessons.put(lesson.id(), lesson);
        }

        // Dars xaridida ham darsning kursi kimniki ekanini bilish kerak
        // (o'z kursini sotib olishni taqiqlash uchun).
        List<Long> lessonCourseIds = new ArrayList<>();
        for (LessonInfo l : lessons.values()) {
            if (!seenCourses.contains(l.courseId())) {
                lessonCourseIds.add(l.courseId());
            }
        }
        Map<Long, CourseDto> lessonCourses = coursesById(lessonCourseIds);
        lessonCourses.putAll(courses);

        // Allaqachon egalik qilinayotgan narsani qayta sotib bo'lmaydi.
        Set<Long> ownedCourses = enrollments.ownedCourses(claims.id(), courseIds);
        Set<Long> ownedLessons = enrollments.ownedLessons(claims.id(), lessonIds);

        OrderDto order = new OrderDto();
        order.userId = claims.id();
        order.status = "paid"; // mock to'lov
        order.paymentMethod = paymentMethod;

        for (int i = 0; i < courseIds.size(); i++) {
            long id = courseIds.get(i);
            CourseDto course = courses.get(id);
            if (course == null || !course.isPublished) {
                v.addError("items[%d].course_id".formatted(i), "course does not exist");
                continue;
            }
            if (course.instructor.id == claims.id()) {
                v.addError("items[%d].course_id".formatted(i), "you cannot purchase your own course");
                continue;
            }
            if (ownedCourses.contains(id)) {
                v.addError("items[%d].course_id".formatted(i), "you already own this course");
                continue;
            }
            OrderItemDto item = new OrderItemDto();
            item.courseId = course.id;
            item.courseTitle = course.title;
            item.instructor = course.instructor.name;
            item.thumbnailColor = UiDefaults.thumbnailColor(course.id);
            item.price = course.price;
            order.items.add(item);
        }

        for (int i = 0; i < lessonIds.size(); i++) {
            long id = lessonIds.get(i);
            LessonInfo lesson = lessons.get(id);
            if (lesson == null) {
                v.addError("items[%d].lesson_id".formatted(i), "lesson does not exist");
                continue;
            }
            CourseDto lessonCourse = lessonCourses.get(lesson.courseId());
            if (lessonCourse != null && lessonCourse.instructor.id == claims.id()) {
                v.addError("items[%d].lesson_id".formatted(i), "you cannot purchase a lesson from your own course");
                continue;
            }
            if (ownedLessons.contains(id)) {
                v.addError("items[%d].lesson_id".formatted(i), "you already own this lesson");
                continue;
            }
            OrderItemDto item = new OrderItemDto();
            item.lessonId = lesson.id();
            item.courseTitle = lesson.courseTitle();
            item.thumbnailColor = UiDefaults.thumbnailColor(lesson.courseId());
            item.price = lesson.price();
            order.items.add(item);
        }

        v.throwIfInvalid();

        orders.insert(order);
        order.finalizeView();

        // To'lov o'tdi — endi kirish beriladi. Kurs: enrollment + barcha
        // darslar. Alohida dars: faqat o'sha dars.
        for (Long id : courseIds) {
            EnrollmentRepository.InsertResult result = enrollments.insert(claims.id(), id);
            service.grantCourseAccess(claims.id(), id);
            if (result.isNew()) {
                service.markEnrolled(id);
            }
        }

        for (Long id : lessonIds) {
            LessonInfo lesson = lessons.get(id);
            enrollments.grantLessonAccess(claims.id(), lesson.courseId(), List.of(lesson.id()));
        }

        service.notify(claims.id(), "system", "Purchase successful",
                "Your order #%s has been paid. %d item(s) are now available in your library."
                        .formatted(order.publicId, order.items.size()));

        return ResponseEntity.status(HttpStatus.CREATED).body(Map.of("order", order));
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

    private static void validatePage(int page, int pageSize) {
        Validator v = new Validator();
        v.check(page > 0, "page", "must be greater than zero");
        v.check(pageSize > 0 && pageSize <= 100, "pageSize", "must be between 1 and 100");
        v.throwIfInvalid();
    }

    private Map<Long, CourseDto> coursesById(List<Long> ids) {
        Map<Long, CourseDto> map = new HashMap<>();
        for (CourseDto c : courseApi.coursesByIds(ids)) {
            map.put(c.id, c);
        }
        return map;
    }
}
