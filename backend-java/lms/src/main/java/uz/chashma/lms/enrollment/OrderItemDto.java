package uz.chashma.lms.enrollment;

import com.fasterxml.jackson.annotation.JsonIgnore;

/** Frontend OrderItem tipi bilan bir xil JSON. */
class OrderItemDto {
    @JsonIgnore
    public Long courseId;
    @JsonIgnore
    public Long lessonId;
    public String courseTitle = "";
    public String instructor = "";
    public String thumbnailColor = "";
    public double price;
}
