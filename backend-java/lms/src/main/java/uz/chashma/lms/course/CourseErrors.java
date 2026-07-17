package uz.chashma.lms.course;

/** Go data paketidagi domen xatolarining ekvivalentlari. */
final class CourseErrors {

    static class DuplicateSlugException extends RuntimeException {
    }

    static class InvalidParentException extends RuntimeException {
    }

    static class MaxDepthExceededException extends RuntimeException {
    }

    static class InvalidCourseException extends RuntimeException {
    }

    private CourseErrors() {
    }
}
