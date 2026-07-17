package uz.chashma.lms.shared.security;

public class InvalidTokenException extends RuntimeException {
    public InvalidTokenException() {
        super("invalid or expired token");
    }
}
