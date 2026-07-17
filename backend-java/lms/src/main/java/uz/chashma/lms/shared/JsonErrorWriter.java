package uz.chashma.lms.shared;

import jakarta.servlet.http.HttpServletResponse;
import org.springframework.http.MediaType;
import tools.jackson.databind.json.JsonMapper;

import java.io.IOException;
import java.util.Map;

/**
 * Security filterlari MVC'dan tashqarida ishlaydi — xato envelope'ini
 * ({"error": ...}) qo'lda yozish uchun yordamchi.
 */
public final class JsonErrorWriter {

    private static final JsonMapper MAPPER = JsonMapper.builder().build();

    private JsonErrorWriter() {
    }

    public static void write(HttpServletResponse response, int status, Object message) throws IOException {
        response.setStatus(status);
        response.setContentType(MediaType.APPLICATION_JSON_VALUE);
        MAPPER.writeValue(response.getOutputStream(), Map.of("error", message));
    }
}
