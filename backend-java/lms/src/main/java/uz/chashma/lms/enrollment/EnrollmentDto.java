package uz.chashma.lms.enrollment;

import java.time.Instant;

/** Frontend Enrollment JSON'i (camelCase). */
class EnrollmentDto {
    public long id;
    public Instant createdAt;
    public long userId;
    public long courseId;
}
