package uz.chashma.lms.course;

import java.util.regex.Pattern;

/** Go data.Slugify porti: "My Category!" -> "my-category". */
final class Slugs {

    private static final Pattern NON_SLUG_CHARS = Pattern.compile("[^a-z0-9]+");

    private Slugs() {
    }

    static String slugify(String s) {
        s = s.trim().toLowerCase();
        s = NON_SLUG_CHARS.matcher(s).replaceAll("-");
        return trim(s, '-');
    }

    private static String trim(String s, char c) {
        int start = 0;
        int end = s.length();
        while (start < end && s.charAt(start) == c) {
            start++;
        }
        while (end > start && s.charAt(end - 1) == c) {
            end--;
        }
        return s.substring(start, end);
    }
}
