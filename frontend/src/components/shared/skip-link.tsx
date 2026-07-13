"use client";

import { useT } from "@/providers/locale-provider";

/**
 * Keyboard/screen-reader shortcut that jumps straight to the page's main
 * content. Visually hidden until focused (first Tab press).
 */
export function SkipLink() {
  const t = useT();
  return (
    <a
      href="#main-content"
      className="sr-only rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-[100]"
    >
      {t("a11y.skipToContent")}
    </a>
  );
}
