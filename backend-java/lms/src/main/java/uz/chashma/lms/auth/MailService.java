package uz.chashma.lms.auth;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.mail.SimpleMailMessage;
import org.springframework.mail.javamail.JavaMailSender;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;

/**
 * Parol tiklash emaili. SMTP sozlanmagan bo'lsa (SMTP_HOST bo'sh) token
 * dev-logga yoziladi — lokal muhitda compose'dagi Mailpit ishlatiladi.
 */
@Service
class MailService {

    private static final Logger log = LoggerFactory.getLogger(MailService.class);

    private final JavaMailSender sender;
    private final String smtpHost;
    private final String from;
    private final String frontendUrl;

    MailService(JavaMailSender sender,
                @Value("${spring.mail.host:}") String smtpHost,
                @Value("${lms.mail.from}") String from,
                @Value("${lms.frontend-url}") String frontendUrl) {
        this.sender = sender;
        this.smtpHost = smtpHost;
        this.from = from;
        this.frontendUrl = frontendUrl;
    }

    /**
     * Async — SMTP sekin bo'lsa ham forgot-password javobi kechikmaydi
     * (javob vaqti orqali email mavjudligini bilib bo'lmasligi ham kerak).
     */
    @Async
    void sendPasswordReset(String to, String token) {
        String link = frontendUrl + "/reset-password?token=" + token;

        if (smtpHost == null || smtpHost.isBlank()) {
            log.info("SMTP sozlanmagan — parol tiklash havolasi: email={} link={}", to, link);
            return;
        }

        SimpleMailMessage message = new SimpleMailMessage();
        message.setFrom(from);
        message.setTo(to);
        message.setSubject("Reset your LearnHub password");
        message.setText("""
                Hi,

                We received a request to reset the password for your account.
                Open the link below to choose a new password (valid for 45 minutes):

                %s

                If you didn't request this, you can safely ignore this email.
                """.formatted(link));

        try {
            sender.send(message);
            log.info("password reset email sent: email={}", to);
        } catch (Exception e) {
            // Xato javobga chiqmaydi (email mavjudligi oshkor bo'lmasin).
            log.error("password reset email failed: email={}", to, e);
        }
    }
}
