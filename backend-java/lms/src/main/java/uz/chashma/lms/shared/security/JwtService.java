package uz.chashma.lms.shared.security;

import io.jsonwebtoken.Claims;
import io.jsonwebtoken.JwtException;
import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.security.Keys;

import javax.crypto.SecretKey;
import java.nio.charset.StandardCharsets;
import java.time.Duration;
import java.time.Instant;
import java.util.Date;

/**
 * Go pkg/auth/jwt.go ekvivalenti — format bir xil: HS256, issuer
 * lms.chashma.uz, sub=userId (string), role claim. Shu tufayli Go
 * yaratgan tokenlar Java'da ham o'qiladi (va aksincha).
 */
public final class JwtService {

    public static final String ISSUER = "lms.chashma.uz";

    private final SecretKey key;
    private final Duration ttl;

    public JwtService(String secret, Duration ttl) {
        // jjwt HS256 uchun kamida 256-bit (32 bayt) kalit talab qiladi.
        this.key = Keys.hmacShaKeyFor(secret.getBytes(StandardCharsets.UTF_8));
        this.ttl = ttl;
    }

    public String newToken(long userId, String role) {
        Instant now = Instant.now();
        return Jwts.builder()
                .subject(Long.toString(userId))
                .issuer(ISSUER)
                .issuedAt(Date.from(now))
                .expiration(Date.from(now.plus(ttl)))
                .claim("role", role)
                .signWith(key, Jwts.SIG.HS256)
                .compact();
    }

    public JwtClaims parse(String token) {
        try {
            Claims claims = Jwts.parser()
                    .verifyWith(key)
                    .requireIssuer(ISSUER)
                    .build()
                    .parseSignedClaims(token)
                    .getPayload();

            if (claims.getExpiration() == null) {
                throw new InvalidTokenException();
            }

            long userId = Long.parseLong(claims.getSubject());
            if (userId < 1) {
                throw new InvalidTokenException();
            }

            return new JwtClaims(userId, claims.get("role", String.class));
        } catch (JwtException | IllegalArgumentException e) {
            throw new InvalidTokenException();
        }
    }
}
