package uz.chashma.lms.shared;

public class NotPermittedException extends RuntimeException {
    public NotPermittedException() {
        super("your user account doesn't have the necessary permissions to access this resource");
    }

    public NotPermittedException(String message) {
        super(message);
    }
}
