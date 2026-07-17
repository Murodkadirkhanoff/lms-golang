package uz.chashma.lms.shared;

import java.util.Map;

public class ValidationException extends RuntimeException {

    private final Map<String, String> errors;

    public ValidationException(Map<String, String> errors) {
        super("validation failed: " + errors);
        this.errors = errors;
    }

    public Map<String, String> errors() {
        return errors;
    }
}
