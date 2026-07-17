package uz.chashma.lms.shared;

import java.util.ArrayList;
import java.util.List;

/** Go jsonutil.ReadIDList ekvivalenti: "1,2,3" -> [1, 2, 3]. */
public final class Ids {

    private Ids() {
    }

    public static List<Long> parse(String csv) {
        List<Long> ids = new ArrayList<>();
        if (csv == null || csv.isBlank()) {
            return ids;
        }
        for (String part : csv.split(",")) {
            try {
                long id = Long.parseLong(part.trim());
                if (id > 0) {
                    ids.add(id);
                }
            } catch (NumberFormatException ignored) {
                // Go ham yaroqsiz bo'laklarni indamay tashlab ketadi
            }
        }
        return ids;
    }
}
