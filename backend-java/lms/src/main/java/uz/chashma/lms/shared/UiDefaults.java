package uz.chashma.lms.shared;

/**
 * Go pkg/uidefaults ekvivalenti: frontend talab qiladigan, lekin DB'da
 * saqlanmaydigan UI maydonlar (avatarColor, thumbnailColor) uchun
 * deterministik defaultlar. Palitra frontend Tailwind klasslariga mos.
 */
public final class UiDefaults {

    private static final String[] PALETTE = {
            "bg-indigo-200",
            "bg-amber-200",
            "bg-rose-200",
            "bg-emerald-200",
            "bg-sky-200",
            "bg-fuchsia-200",
    };

    private UiDefaults() {
    }

    public static String avatarColor(long id) {
        return PALETTE[(int) (id % PALETTE.length)];
    }

    public static String thumbnailColor(long id) {
        return PALETTE[(int) ((id + 3) % PALETTE.length)];
    }
}
