package uz.chashma.lms.auth;

import org.springframework.jdbc.core.simple.JdbcClient;
import org.springframework.stereotype.Repository;

import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.security.SecureRandom;
import java.time.Duration;
import java.time.Instant;
import java.time.OffsetDateTime;
import java.time.ZoneOffset;
import java.util.Optional;

/**
 * Go auth-service/internal/data/tokens.go porti: DB'da faqat SHA-256 hash
 * saqlanadi, plaintext (Base32) foydalanuvchiga yuboriladi.
 */
@Repository
class PasswordResetTokenRepository {

    private static final char[] BASE32_ALPHABET = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567".toCharArray();

    private final JdbcClient jdbc;
    private final SecureRandom random = new SecureRandom();

    PasswordResetTokenRepository(JdbcClient jdbc) {
        this.jdbc = jdbc;
    }

    /** Yangi token yaratib saqlaydi, plaintext'ni qaytaradi. */
    String create(long userId, Duration ttl) {
        byte[] randomBytes = new byte[16];
        random.nextBytes(randomBytes);

        String plaintext = base32NoPadding(randomBytes);

        jdbc.sql("""
                INSERT INTO auth.password_reset_tokens (hash, user_id, expiry)
                VALUES (:hash, :userId, :expiry)
                """)
                .param("hash", sha256(plaintext))
                .param("userId", userId)
                .param("expiry", OffsetDateTime.ofInstant(Instant.now().plus(ttl), ZoneOffset.UTC))
                .update();

        return plaintext;
    }

    Optional<Long> userIdForToken(String plaintext) {
        return jdbc.sql("""
                SELECT user_id
                FROM auth.password_reset_tokens
                WHERE hash = :hash AND expiry > NOW()
                """)
                .param("hash", sha256(plaintext))
                .query(Long.class)
                .optional();
    }

    void deleteAllForUser(long userId) {
        jdbc.sql("DELETE FROM auth.password_reset_tokens WHERE user_id = :userId")
                .param("userId", userId)
                .update();
    }

    private static byte[] sha256(String s) {
        try {
            return MessageDigest.getInstance("SHA-256").digest(s.getBytes(java.nio.charset.StandardCharsets.UTF_8));
        } catch (NoSuchAlgorithmException e) {
            throw new IllegalStateException(e);
        }
    }

    /** Go base32.StdEncoding (paddingsiz) bilan bir xil natija. */
    private static String base32NoPadding(byte[] data) {
        StringBuilder sb = new StringBuilder((data.length * 8 + 4) / 5);
        int buffer = 0;
        int bits = 0;
        for (byte b : data) {
            buffer = (buffer << 8) | (b & 0xFF);
            bits += 8;
            while (bits >= 5) {
                sb.append(BASE32_ALPHABET[(buffer >> (bits - 5)) & 31]);
                bits -= 5;
            }
        }
        if (bits > 0) {
            sb.append(BASE32_ALPHABET[(buffer << (5 - bits)) & 31]);
        }
        return sb.toString();
    }
}
