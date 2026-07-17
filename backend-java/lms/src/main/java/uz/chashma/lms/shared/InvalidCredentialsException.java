package uz.chashma.lms.shared;

public class InvalidCredentialsException extends RuntimeException {
    public InvalidCredentialsException() {
        super("invalid authentication credentials");
    }
}
