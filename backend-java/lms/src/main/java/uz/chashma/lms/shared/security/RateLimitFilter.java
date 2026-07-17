package uz.chashma.lms.shared.security;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.http.HttpStatus;
import org.springframework.web.filter.OncePerRequestFilter;
import uz.chashma.lms.shared.JsonErrorWriter;

import java.io.IOException;
import java.util.Map;
import java.util.Set;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicLong;

/**
 * Auth endpointlari uchun IP bo'yicha token-bucket rate limiter
 * (brute-force himoyasi). Go greenlight'dagi rateLimit middleware uslubi:
 * har IP'ga rps tezlikda to'ladigan, burst sig'imli bucket.
 * In-memory — bitta instance uchun yetarli; k8s'da ko'p replica bo'lsa
 * har pod o'z hisobini yuritadi (limit amalda replica soniga ko'payadi).
 */
public class RateLimitFilter extends OncePerRequestFilter {

    /** Faqat shu yo'llar cheklanadi: register, login, parol tiklash. */
    private static final Set<String> LIMITED_PATHS = Set.of(
            "/v1/users",
            "/v1/tokens/authentication",
            "/v1/tokens/password-reset",
            "/v1/users/password");

    private static final long CLEANUP_INTERVAL_NANOS = 60_000_000_000L; // 1 min
    private static final long STALE_AFTER_NANOS = 180_000_000_000L;     // 3 min

    private final boolean enabled;
    private final double rps;
    private final int burst;

    private final ConcurrentHashMap<String, Bucket> buckets = new ConcurrentHashMap<>();
    private final AtomicLong lastCleanup = new AtomicLong(System.nanoTime());

    public RateLimitFilter(boolean enabled, double rps, int burst) {
        this.enabled = enabled;
        this.rps = rps;
        this.burst = burst;
    }

    private static final class Bucket {
        double tokens;
        long lastSeenNanos;

        Bucket(double tokens, long now) {
            this.tokens = tokens;
            this.lastSeenNanos = now;
        }
    }

    @Override
    protected boolean shouldNotFilter(HttpServletRequest request) {
        return !enabled || !LIMITED_PATHS.contains(request.getRequestURI());
    }

    @Override
    protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response,
                                    FilterChain chain) throws ServletException, IOException {
        long now = System.nanoTime();
        cleanupIfDue(now);

        Bucket bucket = buckets.computeIfAbsent(clientIp(request), ip -> new Bucket(burst, now));

        boolean allowed;
        synchronized (bucket) {
            double refilled = bucket.tokens + (now - bucket.lastSeenNanos) / 1_000_000_000.0 * rps;
            bucket.tokens = Math.min(burst, refilled);
            bucket.lastSeenNanos = now;
            allowed = bucket.tokens >= 1;
            if (allowed) {
                bucket.tokens -= 1;
            }
        }

        if (!allowed) {
            JsonErrorWriter.write(response, HttpStatus.TOO_MANY_REQUESTS.value(), "rate limit exceeded");
            return;
        }

        chain.doFilter(request, response);
    }

    /** Gateway ortida real IP X-Forwarded-For'ning birinchi qiymatida keladi. */
    private static String clientIp(HttpServletRequest request) {
        String forwarded = request.getHeader("X-Forwarded-For");
        if (forwarded != null && !forwarded.isBlank()) {
            return forwarded.split(",")[0].trim();
        }
        return request.getRemoteAddr();
    }

    /** Eski IP yozuvlarini davriy tozalash (xotira o'smasligi uchun). */
    private void cleanupIfDue(long now) {
        long last = lastCleanup.get();
        if (now - last < CLEANUP_INTERVAL_NANOS || !lastCleanup.compareAndSet(last, now)) {
            return;
        }
        for (Map.Entry<String, Bucket> entry : buckets.entrySet()) {
            synchronized (entry.getValue()) {
                if (now - entry.getValue().lastSeenNanos > STALE_AFTER_NANOS) {
                    buckets.remove(entry.getKey());
                }
            }
        }
    }
}
