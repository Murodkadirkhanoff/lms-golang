package uz.chashma.lms.course;

import com.fasterxml.jackson.annotation.JsonIgnore;

import java.time.Instant;

class QuizAttemptDto {
    public long id;
    public Instant createdAt;
    @JsonIgnore
    public long userId;
    @JsonIgnore
    public long courseId;
    public int score;
}
