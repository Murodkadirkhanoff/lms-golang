package uz.chashma.lms.shared;

import org.springframework.boot.context.properties.ConfigurationProperties;

import java.time.Duration;

@ConfigurationProperties(prefix = "lms")
public record LmsProperties(String env, Jwt jwt, String corsTrustedOrigins, RateLimit rateLimit) {

    public record Jwt(String secret, Duration ttl) {
    }

    public record RateLimit(boolean enabled, double rps, int burst) {
    }
}
