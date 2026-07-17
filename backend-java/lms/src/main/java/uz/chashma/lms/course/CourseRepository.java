package uz.chashma.lms.course;

import org.springframework.dao.DataAccessException;
import org.springframework.jdbc.core.RowMapper;
import org.springframework.jdbc.core.simple.JdbcClient;
import org.springframework.stereotype.Repository;
import org.springframework.transaction.annotation.Transactional;
import uz.chashma.lms.course.api.CourseApi.CourseStats;
import uz.chashma.lms.course.api.CourseApi.LessonInfo;
import uz.chashma.lms.course.api.CourseDto;
import uz.chashma.lms.course.api.LessonDto;
import uz.chashma.lms.course.api.ModuleDto;
import uz.chashma.lms.shared.EditConflictException;
import uz.chashma.lms.shared.NotFoundException;
import uz.chashma.lms.shared.PgErrors;

import java.time.OffsetDateTime;
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

/** Go course-service/internal/data/courses.go porti (SQL 1:1). */
@Repository
class CourseRepository {

    private final JdbcClient jdbc;

    CourseRepository(JdbcClient jdbc) {
        this.jdbc = jdbc;
    }

    record CourseFilters(String search, String categorySlug, String sort, int page, int pageSize,
                         List<Long> ids, long instructorId, boolean includeUnpublished) {
    }

    record ListResult(List<CourseDto> courses, int total) {
    }

    // Kurs ro'yxati/detali uchun umumiy ustunlar va aggregatlar.
    private static final String LIST_SELECT = """
            SELECT count(*) OVER() AS total, c.id, c.created_at, c.slug, c.title, c.description,
                   c.thumbnail_url, c.category_id, COALESCE(cat.slug, '') AS category_slug,
                   c.lang, c.price, c.is_published, c.instructor_id, c.student_count,
                   COALESCE(agg.total_lessons, 0) AS total_lessons, COALESCE(agg.total_seconds, 0) AS total_seconds,
                   COALESCE(rv.avg_rating, 0) AS avg_rating, COALESCE(rv.rating_count, 0) AS rating_count
            FROM course.courses c
            LEFT JOIN course.categories cat ON cat.id = c.category_id AND cat.deleted_at IS NULL
            LEFT JOIN LATERAL (
                SELECT count(l.id) AS total_lessons,
                       COALESCE(sum(l.duration_seconds), 0) AS total_seconds
                FROM course.modules m
                JOIN course.lessons l ON l.module_id = m.id
                WHERE m.course_id = c.id
            ) agg ON true
            LEFT JOIN LATERAL (
                SELECT round(avg(r.rating)::numeric, 1) AS avg_rating,
                       count(r.id) AS rating_count
                FROM course.reviews r
                WHERE r.course_id = c.id
            ) rv ON true
            """;

    private static final RowMapper<CourseDto> COURSE_MAPPER = (rs, rowNum) -> {
        CourseDto c = new CourseDto();
        c.id = rs.getLong("id");
        c.createdAt = rs.getObject("created_at", OffsetDateTime.class).toInstant();
        c.slug = rs.getString("slug");
        c.title = rs.getString("title");
        c.description = rs.getString("description");
        c.thumbnailUrl = rs.getString("thumbnail_url");
        long categoryId = rs.getLong("category_id");
        c.categoryId = rs.wasNull() ? null : categoryId;
        c.category = rs.getString("category_slug");
        c.lang = rs.getString("lang");
        c.price = rs.getDouble("price");
        c.isPublished = rs.getBoolean("is_published");
        c.instructorId = rs.getLong("instructor_id");
        c.studentCount = rs.getInt("student_count");
        c.totalLessons = rs.getInt("total_lessons");
        c.totalDurationMinutes = rs.getInt("total_seconds") / 60;
        c.rating = rs.getDouble("avg_rating");
        c.ratingCount = rs.getInt("rating_count");
        return c;
    };

    ListResult list(CourseFilters filters) {
        String orderBy = switch (filters.sort()) {
            case "popular" -> "c.student_count DESC, c.created_at DESC, c.id DESC";
            case "price-asc" -> "c.price ASC, c.id DESC";
            case "price-desc" -> "c.price DESC, c.id DESC";
            default -> "c.created_at DESC, c.id DESC";
        };

        StringBuilder sql = new StringBuilder(LIST_SELECT);
        Map<String, Object> params = new HashMap<>();

        sql.append("""
                WHERE c.deleted_at IS NULL
                  AND (:search = '' OR c.title ILIKE '%' || :search || '%' OR c.description ILIKE '%' || :search || '%')
                  AND (:category = '' OR c.category_id IN (
                        SELECT id FROM course.categories
                        WHERE slug = :category OR parent_id = (SELECT id FROM course.categories WHERE slug = :category)
                  ))
                """);
        params.put("search", filters.search());
        params.put("category", filters.categorySlug());

        if (filters.ids() != null && !filters.ids().isEmpty()) {
            sql.append(" AND c.id IN (:ids)\n");
            params.put("ids", filters.ids());
        }
        if (filters.instructorId() != 0) {
            sql.append(" AND c.instructor_id = :instructorId\n");
            params.put("instructorId", filters.instructorId());
        }
        if (!filters.includeUnpublished()) {
            sql.append(" AND c.is_published = true\n");
        }

        sql.append(" ORDER BY ").append(orderBy).append(" LIMIT :limit OFFSET :offset");
        params.put("limit", filters.pageSize());
        params.put("offset", (filters.page() - 1) * filters.pageSize());

        var totalHolder = new int[1];
        List<CourseDto> courses = jdbc.sql(sql.toString())
                .params(params)
                .query((rs, rowNum) -> {
                    totalHolder[0] = rs.getInt("total");
                    return COURSE_MAPPER.mapRow(rs, rowNum);
                })
                .list();

        return new ListResult(courses, totalHolder[0]);
    }

    /** Bitta kursni modules[].lessons[] bilan to'liq qaytaradi. */
    Optional<CourseDto> findByIdOrSlug(String idOrSlug) {
        long id = 0;
        try {
            id = Long.parseLong(idOrSlug);
        } catch (NumberFormatException ignored) {
        }
        boolean byId = id > 0;

        String sql = LIST_SELECT + " WHERE c.deleted_at IS NULL AND "
                + (byId ? "c.id = :arg" : "c.slug = :arg");

        Optional<CourseDto> course = jdbc.sql(sql)
                .param("arg", byId ? id : idOrSlug)
                .query(COURSE_MAPPER)
                .optional();

        course.ifPresent(c -> c.modules = modulesForCourse(c.id));

        return course;
    }

    List<ModuleDto> modulesForCourse(long courseId) {
        List<ModuleDto> modules = jdbc.sql("""
                SELECT id, title, position
                FROM course.modules
                WHERE course_id = :courseId
                ORDER BY position, id
                """)
                .param("courseId", courseId)
                .query((rs, rowNum) -> {
                    ModuleDto m = new ModuleDto();
                    m.id = rs.getLong("id");
                    m.title = rs.getString("title");
                    m.position = rs.getInt("position");
                    return m;
                })
                .list();

        Map<Long, ModuleDto> byId = new LinkedHashMap<>();
        for (ModuleDto m : modules) {
            byId.put(m.id, m);
        }

        jdbc.sql("""
                SELECT l.id, l.module_id, l.title, l.type, l.content_url, l.content,
                       l.duration_seconds, l.position, l.price, l.is_free
                FROM course.lessons l
                JOIN course.modules m ON m.id = l.module_id
                WHERE m.course_id = :courseId
                ORDER BY l.position, l.id
                """)
                .param("courseId", courseId)
                .query((rs, rowNum) -> {
                    LessonDto l = new LessonDto();
                    l.id = rs.getLong("id");
                    l.title = rs.getString("title");
                    l.type = rs.getString("type");
                    l.contentUrl = rs.getString("content_url");
                    l.content = rs.getString("content");
                    l.durationSeconds = rs.getInt("duration_seconds");
                    l.position = rs.getInt("position");
                    l.price = rs.getDouble("price");
                    l.isFree = rs.getBoolean("is_free");
                    ModuleDto module = byId.get(rs.getLong("module_id"));
                    if (module != null) {
                        module.lessons.add(l);
                    }
                    return l;
                })
                .list();

        return modules;
    }

    /** Kurs + modules + lessons'ni bitta tranzaksiyada yozadi. */
    @Transactional
    void insert(CourseDto course) {
        // Slug band bo'lsa -2, -3... qo'shib ketamiz.
        String base = course.slug;
        for (int i = 2; ; i++) {
            boolean exists = jdbc.sql("SELECT EXISTS(SELECT 1 FROM course.courses WHERE slug = :slug)")
                    .param("slug", course.slug)
                    .query(Boolean.class)
                    .single();
            if (!exists) {
                break;
            }
            course.slug = base + "-" + i;
        }

        try {
            jdbc.sql("""
                    INSERT INTO course.courses (title, slug, description, thumbnail_url, instructor_id, category_id, lang, price, is_published)
                    VALUES (:title, :slug, :description, :thumbnailUrl, :instructorId, :categoryId, :lang, :price, :isPublished)
                    RETURNING id, created_at, version
                    """)
                    .param("title", course.title)
                    .param("slug", course.slug)
                    .param("description", course.description)
                    .param("thumbnailUrl", course.thumbnailUrl == null ? "" : course.thumbnailUrl)
                    .param("instructorId", course.instructorId)
                    .param("categoryId", course.categoryId)
                    .param("lang", course.lang)
                    .param("price", course.price)
                    .param("isPublished", course.isPublished)
                    .query((org.springframework.jdbc.core.RowCallbackHandler) rs -> {
                        course.id = rs.getLong("id");
                        course.createdAt = rs.getObject("created_at", OffsetDateTime.class).toInstant();
                        course.version = rs.getInt("version");
                    });
        } catch (DataAccessException ex) {
            throw constraintError(ex);
        }

        insertModules(course.id, course.modules);

        // Aggregatlar javob uchun (DB'dan qayta o'qimaymiz).
        int totalSeconds = 0;
        course.totalLessons = 0;
        for (ModuleDto module : course.modules) {
            course.totalLessons += module.lessons.size();
            for (LessonDto lesson : module.lessons) {
                totalSeconds += lesson.durationSeconds;
            }
        }
        course.totalDurationMinutes = totalSeconds / 60;
    }

    /** Kursning asosiy maydonlarini yangilaydi (modules/lessons alohida). */
    void update(CourseDto course) {
        try {
            Optional<Integer> version = jdbc.sql("""
                    UPDATE course.courses
                    SET title = :title, description = :description, thumbnail_url = :thumbnailUrl,
                        category_id = :categoryId, lang = :lang,
                        price = :price, is_published = :isPublished, version = version + 1
                    WHERE id = :id AND version = :version AND deleted_at IS NULL
                    RETURNING version
                    """)
                    .param("title", course.title)
                    .param("description", course.description)
                    .param("thumbnailUrl", course.thumbnailUrl == null ? "" : course.thumbnailUrl)
                    .param("categoryId", course.categoryId)
                    .param("lang", course.lang)
                    .param("price", course.price)
                    .param("isPublished", course.isPublished)
                    .param("id", course.id)
                    .param("version", course.version)
                    .query(Integer.class)
                    .optional();
            course.version = version.orElseThrow(EditConflictException::new);
        } catch (DataAccessException ex) {
            throw constraintError(ex);
        }
    }

    /**
     * Kurs o'quv rejasini butunlay almashtiradi (tahrirlash). Eski darslar
     * o'chib yangi id'lar beriladi — sotib olganlarga kirish enrollment
     * tomonida qayta so'ralganda tiklanadi.
     */
    @Transactional
    void replaceModules(long courseId, List<ModuleDto> modules) {
        jdbc.sql("DELETE FROM course.modules WHERE course_id = :courseId")
                .param("courseId", courseId)
                .update();
        insertModules(courseId, modules);
    }

    private void insertModules(long courseId, List<ModuleDto> modules) {
        if (modules == null) {
            return;
        }
        for (ModuleDto module : modules) {
            module.id = jdbc.sql("""
                    INSERT INTO course.modules (course_id, title, position)
                    VALUES (:courseId, :title, :position) RETURNING id
                    """)
                    .param("courseId", courseId)
                    .param("title", module.title)
                    .param("position", module.position)
                    .query(Long.class)
                    .single();

            for (LessonDto lesson : module.lessons) {
                lesson.id = jdbc.sql("""
                        INSERT INTO course.lessons (module_id, title, type, content_url, content,
                                                    duration_seconds, position, price, is_free)
                        VALUES (:moduleId, :title, :type, :contentUrl, :content,
                                :durationSeconds, :position, :price, :isFree)
                        RETURNING id
                        """)
                        .param("moduleId", module.id)
                        .param("title", lesson.title)
                        .param("type", lesson.type)
                        .param("contentUrl", lesson.contentUrl)
                        .param("content", lesson.content)
                        .param("durationSeconds", lesson.durationSeconds)
                        .param("position", lesson.position)
                        .param("price", lesson.price)
                        .param("isFree", lesson.isFree)
                        .query(Long.class)
                        .single();
            }
        }
    }

    /** Soft-delete (enrollments boshqa modulda — qattiq o'chirish xavfli). */
    void delete(long id) {
        if (id < 1) {
            throw new NotFoundException();
        }
        int updated = jdbc.sql(
                "UPDATE course.courses SET deleted_at = NOW() WHERE id = :id AND deleted_at IS NULL")
                .param("id", id)
                .update();
        if (updated == 0) {
            throw new NotFoundException();
        }
    }

    private static final RowMapper<LessonInfo> LESSON_INFO_MAPPER = (rs, rowNum) -> new LessonInfo(
            rs.getLong(1), rs.getString(2), rs.getDouble(3), rs.getBoolean(4),
            rs.getLong(5), rs.getString(6));

    List<LessonInfo> lessonsForCourses(List<Long> courseIds) {
        if (courseIds.isEmpty()) {
            return List.of();
        }
        return jdbc.sql("""
                SELECT l.id, l.title, l.price, l.is_free, c.id, c.title
                FROM course.lessons l
                JOIN course.modules m ON m.id = l.module_id
                JOIN course.courses c ON c.id = m.course_id
                WHERE c.id IN (:ids) AND c.deleted_at IS NULL
                ORDER BY c.id, m.position, m.id, l.position, l.id
                """)
                .param("ids", courseIds)
                .query(LESSON_INFO_MAPPER)
                .list();
    }

    List<LessonInfo> lessonsByIds(List<Long> ids) {
        if (ids.isEmpty()) {
            return List.of();
        }
        return jdbc.sql("""
                SELECT l.id, l.title, l.price, l.is_free, c.id, c.title
                FROM course.lessons l
                JOIN course.modules m ON m.id = l.module_id
                JOIN course.courses c ON c.id = m.course_id
                WHERE l.id IN (:ids) AND c.deleted_at IS NULL
                """)
                .param("ids", ids)
                .query(LESSON_INFO_MAPPER)
                .list();
    }

    void incrementStudentCount(long courseId) {
        jdbc.sql("UPDATE course.courses SET student_count = student_count + 1 WHERE id = :id")
                .param("id", courseId)
                .update();
    }

    CourseStats stats() {
        return jdbc.sql("""
                SELECT count(*) AS total, count(DISTINCT instructor_id) AS instructors
                FROM course.courses
                WHERE deleted_at IS NULL AND is_published = true
                """)
                .query((rs, rowNum) -> new CourseStats(rs.getInt("total"), rs.getInt("instructors")))
                .single();
    }

    Map<Long, Integer> courseCountsByInstructor(List<Long> ids) {
        Map<Long, Integer> counts = new HashMap<>();
        if (ids.isEmpty()) {
            return counts;
        }
        jdbc.sql("""
                SELECT instructor_id, count(*) AS n
                FROM course.courses
                WHERE instructor_id IN (:ids) AND deleted_at IS NULL
                GROUP BY instructor_id
                """)
                .param("ids", ids)
                .query((org.springframework.jdbc.core.RowCallbackHandler) rs ->
                        counts.put(rs.getLong("instructor_id"), rs.getInt("n")));
        return counts;
    }

    record InstructorStat(long instructorId, int courseCount, int students, double rating) {
    }

    List<InstructorStat> instructorStats() {
        return jdbc.sql("""
                SELECT c.instructor_id, count(*) AS course_count, COALESCE(sum(c.student_count), 0) AS students,
                       COALESCE(round(avg(rv.avg_rating)::numeric, 1), 0) AS rating
                FROM course.courses c
                LEFT JOIN LATERAL (
                    SELECT avg(r.rating) AS avg_rating
                    FROM course.reviews r
                    WHERE r.course_id = c.id
                ) rv ON true
                WHERE c.deleted_at IS NULL AND c.is_published = true
                GROUP BY c.instructor_id
                ORDER BY sum(c.student_count) DESC, count(*) DESC
                """)
                .query((rs, rowNum) -> new InstructorStat(
                        rs.getLong("instructor_id"), rs.getInt("course_count"),
                        rs.getInt("students"), rs.getDouble("rating")))
                .list();
    }

    private static RuntimeException constraintError(DataAccessException ex) {
        String constraint = PgErrors.constraint(ex);
        if (constraint != null) {
            switch (constraint) {
                case "courses_slug_key":
                    return new CourseErrors.DuplicateSlugException();
                case "courses_category_id_fkey":
                    return new CourseErrors.InvalidParentException();
            }
        }
        return ex;
    }
}
