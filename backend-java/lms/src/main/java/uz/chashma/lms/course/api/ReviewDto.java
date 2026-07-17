package uz.chashma.lms.course.api;

import com.fasterxml.jackson.annotation.JsonIgnore;

import java.time.Instant;

/** Frontend Review tipi: user — ism (snapshot), avatarColor deterministik. */
public class ReviewDto {
    public long id;
    public Instant createdAt;
    @JsonIgnore
    public long courseId;
    @JsonIgnore
    public long userId;
    public String user;
    public String avatarColor;
    public int rating;
    public String comment;
}
