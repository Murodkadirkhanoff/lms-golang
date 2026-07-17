package uz.chashma.lms.shared.upload;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.multipart.MultipartFile;
import uz.chashma.lms.shared.BadRequestException;

import java.io.IOException;
import java.io.UncheckedIOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.time.LocalDate;
import java.util.Map;
import java.util.Set;
import java.util.UUID;

/**
 * Video/rasm yuklash (S3 o'rniga lokal disk — docker volume'da saqlanadi,
 * /uploads/** orqali beriladi). Keyin S3'ga ko'chirilsa faqat shu klass
 * o'zgaradi, qaytarilgan URL kontrakti saqlanadi.
 */
@RestController
public class UploadController {

    /** Kengaytma oq ro'yxati — executable/HTML yuklab bo'lmaydi. */
    private static final Map<String, Set<String>> ALLOWED = Map.of(
            "video", Set.of("mp4", "webm", "mov", "m4v"),
            "image", Set.of("jpg", "jpeg", "png", "webp", "gif"));

    private final Path root;
    private final String publicUrl;

    public UploadController(@Value("${lms.uploads.dir}") String dir,
                            @Value("${lms.public-url}") String publicUrl) {
        this.root = Path.of(dir).toAbsolutePath().normalize();
        this.publicUrl = publicUrl.endsWith("/") ? publicUrl.substring(0, publicUrl.length() - 1) : publicUrl;
        try {
            Files.createDirectories(root);
        } catch (IOException e) {
            throw new UncheckedIOException("uploads dir yaratilmadi: " + root, e);
        }
    }

    // POST /v1/uploads?kind=video|image (multipart "file"). Auth talab qilinadi
    // (SecurityConfig). Javob: {"url": "...", "filename": "..."}.
    @PostMapping("/v1/uploads")
    ResponseEntity<Map<String, Object>> upload(@RequestParam("file") MultipartFile file,
                                               @RequestParam(defaultValue = "video") String kind) {
        Set<String> allowedExts = ALLOWED.get(kind);
        if (allowedExts == null) {
            throw new BadRequestException("kind must be one of video, image");
        }
        if (file == null || file.isEmpty()) {
            throw new BadRequestException("file must be provided");
        }

        String original = file.getOriginalFilename() == null ? "" : file.getOriginalFilename();
        String ext = extension(original);
        if (!allowedExts.contains(ext)) {
            throw new BadRequestException(
                    "unsupported %s file type .%s (allowed: %s)".formatted(kind, ext, String.join(", ", allowedExts)));
        }

        LocalDate today = LocalDate.now();
        Path dir = root.resolve("%d/%02d".formatted(today.getYear(), today.getMonthValue()));
        String filename = UUID.randomUUID() + "." + ext;
        try {
            Files.createDirectories(dir);
            file.transferTo(dir.resolve(filename));
        } catch (IOException e) {
            throw new UncheckedIOException("faylni saqlab bo'lmadi", e);
        }

        String url = "%s/uploads/%d/%02d/%s".formatted(publicUrl, today.getYear(), today.getMonthValue(), filename);
        return ResponseEntity.status(HttpStatus.CREATED)
                .body(Map.of("url", url, "filename", original));
    }

    private static String extension(String filename) {
        int dot = filename.lastIndexOf('.');
        return dot < 0 ? "" : filename.substring(dot + 1).toLowerCase();
    }
}
