package uz.chashma.lms.auth;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import uz.chashma.lms.shared.InvalidCredentialsException;
import uz.chashma.lms.shared.Validator;
import uz.chashma.lms.shared.security.JwtService;
import uz.chashma.lms.shared.security.Roles;

import java.nio.charset.StandardCharsets;
import java.time.Duration;
import java.util.Map;
import java.util.Optional;

import static uz.chashma.lms.shared.Validator.byteLength;
import static uz.chashma.lms.shared.Validator.orEmpty;

/**
 * Go auth-service cmd/api/users.go + tokens.go porti:
 * register, login, parol tiklash oqimi.
 */
@RestController
@RequestMapping("/v1")
class AuthController {

    private final UserRepository users;
    private final PasswordResetTokenRepository tokens;
    private final PasswordEncoder passwordEncoder;
    private final JwtService jwtService;
    private final MailService mail;

    AuthController(UserRepository users, PasswordResetTokenRepository tokens,
                   PasswordEncoder passwordEncoder, JwtService jwtService, MailService mail) {
        this.users = users;
        this.tokens = tokens;
        this.passwordEncoder = passwordEncoder;
        this.jwtService = jwtService;
        this.mail = mail;
    }

    record RegisterRequest(String name, String email, String password) {
    }

    // POST /v1/users. Kontrakt bo'yicha darhol token qaytariladi.
    @PostMapping("/users")
    ResponseEntity<Map<String, Object>> register(@RequestBody RegisterRequest input) {
        String name = orEmpty(input.name());
        String email = orEmpty(input.email());
        String password = orEmpty(input.password());

        Validator v = new Validator();
        validateName(v, name);
        validateEmail(v, email);
        validatePassword(v, password);
        v.throwIfInvalid();

        User user = new User();
        user.name = name;
        user.email = email;
        user.role = Roles.STUDENT;
        user.passwordHash = passwordEncoder.encode(password).getBytes(StandardCharsets.UTF_8);

        try {
            users.insert(user);
        } catch (UserRepository.DuplicateEmailException e) {
            v.addError("email", "a user with this email address already exists");
            v.throwIfInvalid();
        }

        String token = jwtService.newToken(user.id, user.role);

        return ResponseEntity.status(HttpStatus.CREATED)
                .body(Map.of("user", UserResponse.from(user), "token", token));
    }

    record LoginRequest(String email, String password) {
    }

    // POST /v1/tokens/authentication (login). Javob AuthResult: {user, token}.
    @PostMapping("/tokens/authentication")
    Map<String, Object> login(@RequestBody LoginRequest input) {
        String email = orEmpty(input.email());
        String password = orEmpty(input.password());

        Validator v = new Validator();
        validateEmail(v, email);
        validatePassword(v, password);
        v.throwIfInvalid();

        User user = users.findByEmail(email).orElseThrow(InvalidCredentialsException::new);

        boolean matches = passwordEncoder.matches(password,
                new String(user.passwordHash, StandardCharsets.UTF_8));
        if (!matches) {
            throw new InvalidCredentialsException();
        }

        String token = jwtService.newToken(user.id, user.role);

        return Map.of("user", UserResponse.from(user), "token", token);
    }

    record ForgotPasswordRequest(String email) {
    }

    // POST /v1/tokens/password-reset. Email mavjudligini javobda oshkor
    // qilmaymiz; havola MailService orqali yuboriladi.
    @PostMapping("/tokens/password-reset")
    Map<String, Object> forgotPassword(@RequestBody ForgotPasswordRequest input) {
        String email = orEmpty(input.email());

        Validator v = new Validator();
        validateEmail(v, email);
        v.throwIfInvalid();

        Optional<User> user = users.findByEmail(email);
        if (user.isPresent()) {
            String token = tokens.create(user.get().id, Duration.ofMinutes(45));
            mail.sendPasswordReset(user.get().email, token);
        }

        return Map.of("message", "if the email address exists, a password reset link will be sent");
    }

    record ResetPasswordRequest(String password, String token) {
    }

    // PUT /v1/users/password. Token forgot-password oqimida yaratilgan.
    @PutMapping("/users/password")
    @Transactional
    Map<String, Object> resetPassword(@RequestBody ResetPasswordRequest input) {
        String password = orEmpty(input.password());
        String token = orEmpty(input.token());

        Validator v = new Validator();
        validatePassword(v, password);
        v.check(!token.isEmpty(), "token", "must be provided");
        v.throwIfInvalid();

        Optional<Long> userId = tokens.userIdForToken(token);
        if (userId.isEmpty()) {
            v.addError("token", "invalid or expired password reset token");
            v.throwIfInvalid();
        }

        User user = users.findById(userId.get())
                .orElseThrow(() -> new IllegalStateException("user for valid reset token not found"));

        user.passwordHash = passwordEncoder.encode(password).getBytes(StandardCharsets.UTF_8);
        users.updatePassword(user);
        tokens.deleteAllForUser(user.id);

        return Map.of("message", "your password has been reset successfully");
    }

    // Validatsiya xabarlari Go data.ValidateUser bilan aynan bir xil.
    static void validateName(Validator v, String name) {
        v.check(!name.isEmpty(), "name", "must be provided");
        v.check(byteLength(name) <= 500, "name", "must not be more than 500 bytes long");
    }

    static void validateEmail(Validator v, String email) {
        v.check(!email.isEmpty(), "email", "must be provided");
        v.check(Validator.matches(email, Validator.EMAIL_RX), "email", "must be a valid email address");
    }

    static void validatePassword(Validator v, String password) {
        v.check(!password.isEmpty(), "password", "must be provided");
        v.check(byteLength(password) >= 8, "password", "must be at least 8 bytes long");
        v.check(byteLength(password) <= 72, "password", "must not be more than 72 bytes long");
    }
}
