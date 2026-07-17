package uz.chashma.lms.course.api;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonInclude;

import java.time.Instant;
import java.util.List;

/**
 * Frontend Course tipi bilan bir xil JSON (Go data.Course porti).
 * enrollment moduli ham shu obyektni javobiga o'zgarishsiz joylaydi.
 */
public class CourseDto {
    public long id;
    public Instant createdAt;
    public String slug;
    public String title;
    public String description;
    public String thumbnailColor;
    /** Yuklangan rasm URL'i; bo'sh bo'lsa frontend gradient rang ko'rsatadi. */
    public String thumbnailUrl = "";
    public Long categoryId;
    public String category = "";
    public String lang;
    public double price;
    public double rating;
    public int ratingCount;
    public int studentCount;
    public boolean isPublished;
    public InstructorDto instructor;

    // Go'da omitempty: ro'yxatda yo'q, detalda bor
    @JsonInclude(JsonInclude.Include.NON_NULL)
    public List<ModuleDto> modules;
    @JsonInclude(JsonInclude.Include.NON_NULL)
    public List<ReviewDto> reviews;

    public int totalLessons;
    public int totalDurationMinutes;

    @JsonIgnore
    public long instructorId;
    @JsonIgnore
    public int version;
}
