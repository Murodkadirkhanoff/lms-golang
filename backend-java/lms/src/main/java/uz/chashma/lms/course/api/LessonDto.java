package uz.chashma.lms.course.api;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonInclude;

public class LessonDto {
    public long id;
    public String title;
    public String type;
    // Go omitempty: bo'sh bo'lsa chiqmaydi (paywall shu maydonlarni bo'shatadi)
    @JsonInclude(JsonInclude.Include.NON_EMPTY)
    public String contentUrl;
    @JsonInclude(JsonInclude.Include.NON_EMPTY)
    public String content;
    public int durationSeconds;
    @JsonIgnore
    public int position;
    public double price;
    public boolean isFree;
    // Paywall: kontent yashirilganda true, aks holda JSON'da chiqmaydi
    @JsonInclude(JsonInclude.Include.NON_NULL)
    public Boolean locked;
}
