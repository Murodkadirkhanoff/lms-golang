export const LOCALES = ["uz", "ru", "en"] as const;

export type Locale = (typeof LOCALES)[number];

export const DEFAULT_LOCALE: Locale = "uz";

export const LOCALE_LABELS: Record<Locale, { native: string; flag: string }> = {
  uz: { native: "O‘zbekcha", flag: "🇺🇿" },
  ru: { native: "Русский", flag: "🇷🇺" },
  en: { native: "English", flag: "🇬🇧" },
};
