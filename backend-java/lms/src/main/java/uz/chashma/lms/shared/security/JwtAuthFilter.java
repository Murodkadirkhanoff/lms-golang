package uz.chashma.lms.shared.security;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.web.filter.OncePerRequestFilter;
import uz.chashma.lms.shared.JsonErrorWriter;

import java.io.IOException;
import java.util.List;

/**
 * Go pkg/middleware.Authenticate ekvivalenti: Authorization header bo'lmasa
 * anonim davom etadi; noto'g'ri/eskirgan token esa darhol 401 qaytaradi.
 */
public class JwtAuthFilter extends OncePerRequestFilter {

    private final JwtService jwtService;

    public JwtAuthFilter(JwtService jwtService) {
        this.jwtService = jwtService;
    }

    @Override
    protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response, FilterChain chain)
            throws ServletException, IOException {
        String header = request.getHeader(HttpHeaders.AUTHORIZATION);
        if (header == null || header.isBlank()) {
            chain.doFilter(request, response);
            return;
        }

        String[] parts = header.split(" ");
        if (parts.length != 2 || !"Bearer".equals(parts[0])) {
            invalidToken(response);
            return;
        }

        JwtClaims claims;
        try {
            claims = jwtService.parse(parts[1]);
        } catch (InvalidTokenException e) {
            invalidToken(response);
            return;
        }

        var authentication = new UsernamePasswordAuthenticationToken(
                new UserPrincipal(claims.userId(), claims.role()),
                null,
                List.of(new SimpleGrantedAuthority("ROLE_" + claims.role())));
        SecurityContextHolder.getContext().setAuthentication(authentication);

        chain.doFilter(request, response);
    }

    private void invalidToken(HttpServletResponse response) throws IOException {
        response.setHeader(HttpHeaders.WWW_AUTHENTICATE, "Bearer");
        JsonErrorWriter.write(response, HttpStatus.UNAUTHORIZED.value(),
                "invalid or missing authentication token");
    }
}
