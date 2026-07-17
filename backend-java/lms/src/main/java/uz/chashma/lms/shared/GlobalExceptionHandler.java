package uz.chashma.lms.shared;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.http.converter.HttpMessageNotReadableException;
import org.springframework.web.HttpRequestMethodNotSupportedException;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.RestControllerAdvice;
import org.springframework.web.method.annotation.MethodArgumentTypeMismatchException;
import org.springframework.web.servlet.resource.NoResourceFoundException;

import java.util.Map;

/**
 * Go pkg/httperr.Responder ekvivalenti: barcha xatolar
 * {"error": string | {field: message}} envelope'ida qaytadi.
 */
@RestControllerAdvice
public class GlobalExceptionHandler {

    private static final Logger log = LoggerFactory.getLogger(GlobalExceptionHandler.class);

    @ExceptionHandler(ValidationException.class)
    public ResponseEntity<Map<String, Object>> failedValidation(ValidationException ex) {
        return error(HttpStatus.UNPROCESSABLE_ENTITY, ex.errors());
    }

    @ExceptionHandler(BadRequestException.class)
    public ResponseEntity<Map<String, Object>> badRequest(BadRequestException ex) {
        return error(HttpStatus.BAD_REQUEST, ex.getMessage());
    }

    @ExceptionHandler(HttpMessageNotReadableException.class)
    public ResponseEntity<Map<String, Object>> unreadableBody(HttpMessageNotReadableException ex) {
        return error(HttpStatus.BAD_REQUEST, "body contains badly-formed JSON");
    }

    // Yaroqsiz path param ({id} raqam emas) Go'da NotFound qaytaradi.
    @ExceptionHandler({NotFoundException.class, NoResourceFoundException.class,
            MethodArgumentTypeMismatchException.class})
    public ResponseEntity<Map<String, Object>> notFound(Exception ex) {
        return error(HttpStatus.NOT_FOUND, "the requested resource could not be found");
    }

    @ExceptionHandler(HttpRequestMethodNotSupportedException.class)
    public ResponseEntity<Map<String, Object>> methodNotAllowed(HttpRequestMethodNotSupportedException ex) {
        String message = "the %s method is not supported for this resource".formatted(ex.getMethod());
        return error(HttpStatus.METHOD_NOT_ALLOWED, message);
    }

    @ExceptionHandler(InvalidCredentialsException.class)
    public ResponseEntity<Map<String, Object>> invalidCredentials(InvalidCredentialsException ex) {
        return error(HttpStatus.UNAUTHORIZED, ex.getMessage());
    }

    @ExceptionHandler(NotPermittedException.class)
    public ResponseEntity<Map<String, Object>> notPermitted(NotPermittedException ex) {
        return error(HttpStatus.FORBIDDEN, ex.getMessage());
    }

    @ExceptionHandler(EditConflictException.class)
    public ResponseEntity<Map<String, Object>> editConflict(EditConflictException ex) {
        return error(HttpStatus.CONFLICT, ex.getMessage());
    }

    @ExceptionHandler(Exception.class)
    public ResponseEntity<Map<String, Object>> serverError(Exception ex) {
        log.error("unhandled exception", ex);
        return error(HttpStatus.INTERNAL_SERVER_ERROR,
                "the server encountered a problem and could not process your request");
    }

    private ResponseEntity<Map<String, Object>> error(HttpStatus status, Object message) {
        return ResponseEntity.status(status).body(Map.of("error", message));
    }
}
