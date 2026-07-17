package uz.chashma.lms.shared.security;

import org.springframework.security.core.Authentication;
import org.springframework.security.core.context.SecurityContextHolder;

/** Go middleware.ContextGetUser ekvivalenti. */
public final class CurrentUser {

    private CurrentUser() {
    }

    /** Anonim so'rovda null qaytaradi. */
    public static UserPrincipal get() {
        Authentication auth = SecurityContextHolder.getContext().getAuthentication();
        if (auth != null && auth.getPrincipal() instanceof UserPrincipal principal) {
            return principal;
        }
        return null;
    }
}
