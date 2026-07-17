package uz.chashma.lms.course;

import com.fasterxml.jackson.annotation.JsonIgnore;

import java.time.Instant;

/** Frontend Category tipi bilan bir xil JSON (camelCase). */
class CategoryDto {
    public long id;
    @JsonIgnore
    public Instant createdAt;
    public String slug;
    public String nameUz;
    public String nameRu;
    public String nameEn;
    public Long parentId;
    // Published kurslar soni (List'da to'ldiriladi; ota kategoriya uchun
    // bolalarniki bilan birga).
    public int courseCount;
    @JsonIgnore
    public int version;
}
