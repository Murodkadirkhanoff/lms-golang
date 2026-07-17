package uz.chashma.lms.shared;

public class NotFoundException extends RuntimeException {
    public NotFoundException() {
        super("the requested resource could not be found");
    }
}
