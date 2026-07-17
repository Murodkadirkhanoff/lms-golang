package uz.chashma.lms.auth;

import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import tools.jackson.databind.PropertyNamingStrategies;
import tools.jackson.databind.annotation.JsonNaming;
import uz.chashma.lms.shared.InvalidCredentialsException;
import uz.chashma.lms.shared.NotFoundException;
import uz.chashma.lms.shared.Validator;
import uz.chashma.lms.shared.security.CurrentUser;

import java.nio.charset.StandardCharsets;
import java.util.Map;

import static uz.chashma.lms.shared.Validator.orEmpty;

/**
 * Settings/Profile sahifasi uchun: ism va parolni o'zgartirish.
 * Email o'zgartirish yo'q — u login identifikatori.
 */
@RestController
@RequestMapping("/v1/me")
class ProfileController {

    private final UserRepository users;
    private final PasswordEncoder passwordEncoder;

    ProfileController(UserRepository users, PasswordEncoder passwordEncoder) {
        this.users = users;
        this.passwordEncoder = passwordEncoder;
    }

    // GET /v1/me — joriy foydalanuvchi (frontend profilni yangilab olish uchun).
    @GetMapping
    Map<String, Object> me() {
        User user = users.findById(CurrentUser.get().id()).orElseThrow(NotFoundException::new);
        return Map.of("user", UserResponse.from(user));
    }

    record UpdateProfileRequest(String name) {
    }

    // PUT /v1/me/profile — hozircha faqat ism.
    @PutMapping("/profile")
    Map<String, Object> updateProfile(@RequestBody UpdateProfileRequest input) {
        String name = orEmpty(input.name());

        Validator v = new Validator();
        AuthController.validateName(v, name);
        v.throwIfInvalid();

        User user = users.findById(CurrentUser.get().id()).orElseThrow(NotFoundException::new);
        user.name = name;
        users.updateName(user);

        return Map.of("user", UserResponse.from(user));
    }

    @JsonNaming(PropertyNamingStrategies.SnakeCaseStrategy.class)
    record ChangePasswordRequest(String currentPassword, String newPassword) {
    }

    // PUT /v1/me/password — joriy parol tasdig'i bilan.
    @PutMapping("/password")
    Map<String, Object> changePassword(@RequestBody ChangePasswordRequest input) {
        String current = orEmpty(input.currentPassword());
        String updated = orEmpty(input.newPassword());

        Validator v = new Validator();
        v.check(!current.isEmpty(), "current_password", "must be provided");
        AuthController.validatePassword(v, updated);
        v.throwIfInvalid();

        User user = users.findById(CurrentUser.get().id()).orElseThrow(NotFoundException::new);

        boolean matches = passwordEncoder.matches(current,
                new String(user.passwordHash, StandardCharsets.UTF_8));
        if (!matches) {
            throw new InvalidCredentialsException();
        }

        user.passwordHash = passwordEncoder.encode(updated).getBytes(StandardCharsets.UTF_8);
        users.updatePassword(user);

        return Map.of("message", "your password has been updated successfully");
    }
}
