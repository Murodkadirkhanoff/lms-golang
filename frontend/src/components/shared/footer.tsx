"use client";

import Link from "next/link";
import { Logo } from "./logo";
import { ROUTES } from "@/constants";
import { useT } from "@/providers/locale-provider";

const groups = [
  {
    titleKey: "footer.platform",
    links: [
      { key: "footer.courses", href: ROUTES.courses },
      { key: "footer.categories", href: ROUTES.categories },
      { key: "footer.teach", href: ROUTES.teach },
      { key: "footer.pricing", href: ROUTES.pricing },
    ],
  },
  {
    titleKey: "footer.company",
    links: [
      { key: "footer.about", href: ROUTES.about },
      { key: "footer.help", href: ROUTES.help },
      { key: "footer.contact", href: ROUTES.contact },
    ],
  },
  {
    titleKey: "footer.legal",
    links: [
      { key: "footer.privacy", href: ROUTES.privacy },
      { key: "footer.terms", href: ROUTES.terms },
    ],
  },
];

export function Footer() {
  const t = useT();
  return (
    <footer className="border-t">
      <div className="mx-auto grid max-w-7xl gap-8 px-6 py-14 md:grid-cols-5">
        <div className="md:col-span-2">
          <Logo />
          <p className="mt-4 max-w-xs text-sm text-muted-foreground">{t("footer.blurb")}</p>
        </div>
        {groups.map((group) => (
          <div key={group.titleKey}>
            <h4 className="text-sm font-semibold">{t(group.titleKey)}</h4>
            <ul className="mt-3 space-y-2 text-sm text-muted-foreground">
              {group.links.map((link) => (
                <li key={link.key}>
                  <Link href={link.href} className="hover:text-foreground">
                    {t(link.key)}
                  </Link>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
      <div className="border-t py-6 text-center text-sm text-muted-foreground">{t("footer.rights")}</div>
    </footer>
  );
}
