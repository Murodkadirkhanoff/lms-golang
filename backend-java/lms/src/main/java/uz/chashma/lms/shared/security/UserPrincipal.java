package uz.chashma.lms.shared.security;

/** JWT'dan olingan foydalanuvchi — SecurityContext'dagi principal. */
public record UserPrincipal(long id, String role) {
}
