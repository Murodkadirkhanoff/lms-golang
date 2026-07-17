package uz.chashma.lms.course;

import tools.jackson.databind.PropertyNamingStrategies;
import tools.jackson.databind.annotation.JsonNaming;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PatchMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import uz.chashma.lms.shared.NotFoundException;
import uz.chashma.lms.shared.Validator;

import java.net.URI;
import java.util.Map;

import static uz.chashma.lms.shared.Validator.byteLength;
import static uz.chashma.lms.shared.Validator.orEmpty;

/** Go course-service cmd/api/categories.go porti. Request kalitlari snake_case. */
@RestController
@RequestMapping("/v1/categories")
class CategoryController {

    private final CategoryRepository categories;

    CategoryController(CategoryRepository categories) {
        this.categories = categories;
    }

    @GetMapping
    Map<String, Object> list() {
        return Map.of("categories", categories.list());
    }

    @JsonNaming(PropertyNamingStrategies.SnakeCaseStrategy.class)
    record CreateCategoryRequest(String nameUz, String nameRu, String nameEn, Long parentId) {
    }

    @PostMapping
    ResponseEntity<Map<String, Object>> create(@RequestBody CreateCategoryRequest input) {
        CategoryDto category = new CategoryDto();
        category.nameUz = orEmpty(input.nameUz());
        category.nameRu = orEmpty(input.nameRu());
        category.nameEn = orEmpty(input.nameEn());
        category.parentId = input.parentId();

        Validator v = new Validator();
        validate(v, category);
        v.throwIfInvalid();

        category.slug = Slugs.slugify(category.nameEn);
        if (category.slug.isEmpty()) {
            v.addError("name_en", "must contain at least one latin letter or digit for slug generation");
            v.throwIfInvalid();
        }

        insertOrUpdate(v, () -> categories.insert(category));

        return ResponseEntity.created(URI.create("/v1/categories/" + category.id))
                .body(Map.of("category", category));
    }

    @GetMapping("/{id}")
    Map<String, Object> show(@PathVariable long id) {
        CategoryDto category = categories.findById(id).orElseThrow(NotFoundException::new);
        return Map.of("category", category);
    }

    @JsonNaming(PropertyNamingStrategies.SnakeCaseStrategy.class)
    record UpdateCategoryRequest(String nameUz, String nameRu, String nameEn, Long parentId) {
    }

    @PatchMapping("/{id}")
    Map<String, Object> update(@PathVariable long id, @RequestBody UpdateCategoryRequest input) {
        CategoryDto category = categories.findById(id).orElseThrow(NotFoundException::new);

        if (input.nameUz() != null) {
            category.nameUz = input.nameUz();
        }
        if (input.nameRu() != null) {
            category.nameRu = input.nameRu();
        }
        if (input.nameEn() != null) {
            category.nameEn = input.nameEn();
        }
        if (input.parentId() != null) {
            category.parentId = input.parentId();
        }

        Validator v = new Validator();
        validate(v, category);
        v.throwIfInvalid();

        if (input.nameEn() != null) {
            category.slug = Slugs.slugify(category.nameEn);
            if (category.slug.isEmpty()) {
                v.addError("name_en", "must contain at least one latin letter or digit for slug generation");
                v.throwIfInvalid();
            }
        }

        insertOrUpdate(v, () -> categories.update(category));

        return Map.of("category", category);
    }

    @DeleteMapping("/{id}")
    Map<String, Object> delete(@PathVariable long id) {
        categories.delete(id);
        return Map.of("message", "category successfully deleted");
    }

    // Go handleCategoryWriteError ekvivalenti.
    private static void insertOrUpdate(Validator v, Runnable write) {
        try {
            write.run();
        } catch (CourseErrors.DuplicateSlugException e) {
            v.addError("name_en", "a category with this name already exists");
            v.throwIfInvalid();
        } catch (CourseErrors.InvalidParentException e) {
            v.addError("parent_id", "parent category does not exist");
            v.throwIfInvalid();
        } catch (CourseErrors.MaxDepthExceededException e) {
            v.addError("parent_id", "category nesting is too deep (max 2 levels)");
            v.throwIfInvalid();
        }
    }

    // Go data.ValidateCategory bilan bir xil xabarlar.
    private static void validate(Validator v, CategoryDto category) {
        v.check(!category.nameUz.isEmpty(), "name_uz", "must be provided");
        v.check(byteLength(category.nameUz) <= 100, "name_uz", "must not be more than 100 bytes long");

        v.check(!category.nameRu.isEmpty(), "name_ru", "must be provided");
        v.check(byteLength(category.nameRu) <= 100, "name_ru", "must not be more than 100 bytes long");

        v.check(!category.nameEn.isEmpty(), "name_en", "must be provided");
        v.check(byteLength(category.nameEn) <= 100, "name_en", "must not be more than 100 bytes long");
    }
}
