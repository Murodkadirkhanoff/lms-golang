package uz.chashma.lms.course;

import org.springframework.dao.DataAccessException;
import org.springframework.jdbc.core.RowCallbackHandler;
import org.springframework.jdbc.core.RowMapper;
import org.springframework.jdbc.core.simple.JdbcClient;
import org.springframework.stereotype.Repository;
import uz.chashma.lms.shared.EditConflictException;
import uz.chashma.lms.shared.NotFoundException;
import uz.chashma.lms.shared.PgErrors;

import java.time.OffsetDateTime;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

/** Go course-service/internal/data/categories.go porti. */
@Repository
class CategoryRepository {

    private final JdbcClient jdbc;

    CategoryRepository(JdbcClient jdbc) {
        this.jdbc = jdbc;
    }

    private static final RowMapper<CategoryDto> MAPPER = (rs, rowNum) -> {
        CategoryDto c = new CategoryDto();
        c.id = rs.getLong("id");
        c.createdAt = rs.getObject("created_at", OffsetDateTime.class).toInstant();
        c.slug = rs.getString("slug");
        c.nameUz = rs.getString("name_uz");
        c.nameRu = rs.getString("name_ru");
        c.nameEn = rs.getString("name_en");
        long parentId = rs.getLong("parent_id");
        c.parentId = rs.wasNull() ? null : parentId;
        c.version = rs.getInt("version");
        return c;
    };

    void insert(CategoryDto category) {
        try {
            jdbc.sql("""
                    INSERT INTO course.categories (slug, name_uz, name_ru, name_en, parent_id)
                    VALUES (:slug, :nameUz, :nameRu, :nameEn, :parentId)
                    RETURNING id, created_at, version
                    """)
                    .param("slug", category.slug)
                    .param("nameUz", category.nameUz)
                    .param("nameRu", category.nameRu)
                    .param("nameEn", category.nameEn)
                    .param("parentId", category.parentId)
                    .query((RowCallbackHandler) rs -> {
                        category.id = rs.getLong("id");
                        category.createdAt = rs.getObject("created_at", OffsetDateTime.class).toInstant();
                        category.version = rs.getInt("version");
                    });
        } catch (DataAccessException ex) {
            throw constraintError(ex);
        }
    }

    Optional<CategoryDto> findById(long id) {
        if (id < 1) {
            return Optional.empty();
        }
        return jdbc.sql("""
                SELECT id, created_at, slug, name_uz, name_ru, name_en, parent_id, version
                FROM course.categories
                WHERE id = :id AND deleted_at IS NULL
                """)
                .param("id", id)
                .query(MAPPER)
                .optional();
    }

    List<CategoryDto> list() {
        List<CategoryDto> categories = jdbc.sql("""
                SELECT c.id, c.created_at, c.slug, c.name_uz, c.name_ru, c.name_en, c.parent_id, c.version,
                       COALESCE(cc.n, 0) AS course_count
                FROM course.categories c
                LEFT JOIN (
                    SELECT category_id, COUNT(*) AS n
                    FROM course.courses
                    WHERE is_published = true AND deleted_at IS NULL AND category_id IS NOT NULL
                    GROUP BY category_id
                ) cc ON cc.category_id = c.id
                WHERE c.deleted_at IS NULL
                ORDER BY depth, c.id
                """)
                .query((rs, rowNum) -> {
                    CategoryDto c = MAPPER.mapRow(rs, rowNum);
                    c.courseCount = rs.getInt("course_count");
                    return c;
                })
                .list();

        // Ota kategoriya soni = o'ziniki + bolalariniki (kurslar leaf'ga biriktiriladi).
        Map<Long, CategoryDto> byId = new HashMap<>();
        for (CategoryDto c : categories) {
            byId.put(c.id, c);
        }
        for (CategoryDto c : categories) {
            if (c.parentId != null) {
                CategoryDto parent = byId.get(c.parentId);
                if (parent != null) {
                    parent.courseCount += c.courseCount;
                }
            }
        }

        return categories;
    }

    void update(CategoryDto category) {
        try {
            Optional<Integer> version = jdbc.sql("""
                    UPDATE course.categories
                    SET slug = :slug, name_uz = :nameUz, name_ru = :nameRu, name_en = :nameEn,
                        parent_id = :parentId, version = version + 1
                    WHERE id = :id AND version = :version
                    RETURNING version
                    """)
                    .param("slug", category.slug)
                    .param("nameUz", category.nameUz)
                    .param("nameRu", category.nameRu)
                    .param("nameEn", category.nameEn)
                    .param("parentId", category.parentId)
                    .param("id", category.id)
                    .param("version", category.version)
                    .query(Integer.class)
                    .optional();
            category.version = version.orElseThrow(EditConflictException::new);
        } catch (DataAccessException ex) {
            throw constraintError(ex);
        }
    }

    void delete(long id) {
        if (id < 1) {
            throw new NotFoundException();
        }
        int deleted = jdbc.sql("DELETE FROM course.categories WHERE id = :id")
                .param("id", id)
                .update();
        if (deleted == 0) {
            throw new NotFoundException();
        }
    }

    private static RuntimeException constraintError(DataAccessException ex) {
        String constraint = PgErrors.constraint(ex);
        if (constraint != null) {
            switch (constraint) {
                case "categories_slug_key":
                    return new CourseErrors.DuplicateSlugException();
                case "categories_parent_id_fkey":
                    return new CourseErrors.InvalidParentException();
                case "max_category_depth":
                    return new CourseErrors.MaxDepthExceededException();
            }
        }
        return ex;
    }
}
