package uz.chashma.lms.shared;

import java.nio.charset.StandardCharsets;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.regex.Pattern;

/**
 * Go pkg/validator ekvivalenti: xatolar {field: message} shaklida yig'iladi
 * va 422 javobda aynan shu ko'rinishda chiqadi. Xabar matnlari Go bilan
 * bir xil bo'lishi shart (frontend shu matnlarni ko'rsatadi).
 */
public class Validator {

    public static final Pattern EMAIL_RX = Pattern.compile(
            "^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$");

    private final Map<String, String> errors = new LinkedHashMap<>();

    public void check(boolean ok, String key, String message) {
        if (!ok) {
            addError(key, message);
        }
    }

    public void addError(String key, String message) {
        errors.putIfAbsent(key, message);
    }

    public boolean valid() {
        return errors.isEmpty();
    }

    public Map<String, String> errors() {
        return errors;
    }

    public void throwIfInvalid() {
        if (!valid()) {
            throw new ValidationException(errors);
        }
    }

    /** Go len() bayt uzunligini o'lchaydi — chegara tekshiruvlari baytda. */
    public static int byteLength(String s) {
        return s.getBytes(StandardCharsets.UTF_8).length;
    }

    public static boolean matches(String s, Pattern pattern) {
        return pattern.matcher(s).matches();
    }

    /** Go validator.PermittedValue ekvivalenti. */
    public static boolean permitted(String value, String... allowed) {
        for (String a : allowed) {
            if (a.equals(value)) {
                return true;
            }
        }
        return false;
    }

    /** null'ni bo'sh satrga aylantiradi — Go'da yo'q maydon "" bo'ladi. */
    public static String orEmpty(String s) {
        return s == null ? "" : s;
    }
}
