"use client";

import Link from "next/link";
import { BookOpen, CreditCard, GraduationCap, LifeBuoy, Search, Settings, Award } from "lucide-react";
import { Card } from "@/components/ui/card";
import { ROUTES } from "@/constants";
import { useT } from "@/providers/locale-provider";

const topics = [
  { icon: GraduationCap, titleKey: "help.t1Title", descKey: "help.t1Desc" },
  { icon: CreditCard, titleKey: "help.t2Title", descKey: "help.t2Desc" },
  { icon: BookOpen, titleKey: "help.t3Title", descKey: "help.t3Desc" },
  { icon: Award, titleKey: "help.t4Title", descKey: "help.t4Desc" },
  { icon: Settings, titleKey: "help.t5Title", descKey: "help.t5Desc" },
  { icon: LifeBuoy, titleKey: "help.t6Title", descKey: "help.t6Desc" },
];

const faqs = [
  { qKey: "help.q1", aKey: "help.a1" },
  { qKey: "help.q2", aKey: "help.a2" },
  { qKey: "help.q3", aKey: "help.a3" },
  { qKey: "help.q4", aKey: "help.a4" },
];

export default function HelpPage() {
  const t = useT();
  return (
    <div className="mx-auto max-w-5xl px-6 py-16">
      <div className="text-center">
        <h1 className="text-4xl font-extrabold tracking-tight">{t("help.title")}</h1>
        <div className="relative mx-auto mt-6 max-w-xl">
          <Search className="pointer-events-none absolute left-4 top-1/2 size-5 -translate-y-1/2 text-muted-foreground" />
          <input
            placeholder={t("help.searchPlaceholder")}
            className="h-12 w-full rounded-full border border-input bg-background pl-12 pr-4 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
          />
        </div>
      </div>

      <div className="mt-14 grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
        {topics.map((tp) => (
          <Card key={tp.titleKey} className="p-6 transition-shadow hover:shadow-md">
            <div className="grid size-11 place-items-center rounded-xl bg-accent text-accent-foreground">
              <tp.icon className="size-5" />
            </div>
            <h2 className="mt-4 font-bold">{t(tp.titleKey)}</h2>
            <p className="mt-1 text-sm text-muted-foreground">{t(tp.descKey)}</p>
          </Card>
        ))}
      </div>

      <section className="mt-16">
        <h2 className="text-2xl font-extrabold">{t("help.popular")}</h2>
        <div className="mt-6 space-y-3">
          {faqs.map((f, i) => (
            <details key={f.qKey} className="group rounded-xl border bg-card p-5" open={i === 0}>
              <summary className="flex cursor-pointer items-center justify-between font-semibold">
                {t(f.qKey)}
                <span className="text-primary transition-transform group-open:rotate-45">+</span>
              </summary>
              <p className="mt-3 text-sm text-muted-foreground">{t(f.aKey)}</p>
            </details>
          ))}
        </div>
      </section>

      <Card className="mt-12 flex flex-col items-center gap-3 p-8 text-center">
        <LifeBuoy className="size-8 text-primary" />
        <h2 className="text-lg font-bold">{t("help.stillTitle")}</h2>
        <p className="text-sm text-muted-foreground">{t("help.stillDesc")}</p>
        <Link href={ROUTES.contact} className="font-semibold text-primary hover:underline">
          {t("help.contactSupport")}
        </Link>
      </Card>
    </div>
  );
}
