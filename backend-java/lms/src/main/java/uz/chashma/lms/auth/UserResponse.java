package uz.chashma.lms.auth;

import java.time.Instant;

/** Frontend User tipi bilan bir xil JSON (camelCase). */
record UserResponse(long id, Instant createdAt, String name, String email, String role) {

    static UserResponse from(User user) {
        return new UserResponse(user.id, user.createdAt, user.name, user.email, user.role);
    }
}
