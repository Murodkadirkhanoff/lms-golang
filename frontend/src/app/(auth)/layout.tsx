"use client";

import Link from "next/link";
import { Logo } from "@/components/shared/logo";
import { ThemeToggle } from "@/components/shared/theme-toggle";
import { LanguageSwitcher } from "@/components/shared/language-switcher";
import { SkipLink } from "@/components/shared/skip-link";
import { ROUTES } from "@/constants";
import { useT } from "@/providers/locale-provider";

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  const t = useT();
  return (
    <div className="grid min-h-screen lg:grid-cols-2">
      <SkipLink />
      {/* Brand side */}
      <div className="relative hidden flex-col justify-between bg-gradient-to-br from-primary to-violet-700 p-12 text-primary-foreground lg:flex">
        <Logo light />
        <div>
          <h2 className="text-4xl font-extrabold leading-tight">{t("auth.brandTitle")}</h2>
          <p className="mt-4 max-w-sm text-indigo-100">{t("auth.brandSubtitle")}</p>
        </div>
        <p className="text-sm text-indigo-200">© 2026 LearnHub</p>
      </div>

      {/* Form side */}
      <main id="main-content" className="relative flex flex-col items-center justify-center p-6">
        <div className="absolute right-4 top-4 flex items-center">
          <LanguageSwitcher />
          <ThemeToggle />
        </div>
        <div className="w-full max-w-sm">
          <div className="mb-8 lg:hidden">
            <Logo />
          </div>
          {children}
          <p className="mt-8 text-center text-xs text-muted-foreground">
            {t("auth.agree")}{" "}
            <Link href={ROUTES.terms} className="underline">
              {t("auth.terms")}
            </Link>{" "}
            {t("auth.and")}{" "}
            <Link href={ROUTES.privacy} className="underline">
              {t("auth.privacy")}
            </Link>
            {t("auth.agreeRules")}
          </p>
        </div>
      </main>
    </div>
  );
}
