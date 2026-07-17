package uz.chashma.lms.shared.security;

public record JwtClaims(long userId, String role) {
}
