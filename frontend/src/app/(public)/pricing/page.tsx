"use client";

import Link from "next/link";
import { Check } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ROUTES } from "@/constants";
import { cn } from "@/lib/utils";
import { useT } from "@/providers/locale-provider";

const plans = [
  {
    nameKey: "pricing.freeName",
    price: "$0",
    periodKey: "pricing.forever",
    descKey: "pricing.freeDesc",
    ctaKey: "pricing.freeCta",
    href: ROUTES.register,
    featureKeys: ["pricing.f_freeCourses", "pricing.f_progress", "pricing.f_community", "pricing.f_devices"],
    highlight: false,
  },
  {
    nameKey: "pricing.proName",
    price: "$19",
    periodKey: "pricing.month",
    descKey: "pricing.proDesc",
    ctaKey: "pricing.proCta",
    href: ROUTES.register,
    featureKeys: [
      "pricing.f_everythingFree",
      "pricing.f_unlimited",
      "pricing.f_certs",
      "pricing.f_downloads",
      "pricing.f_offline",
      "pricing.f_support",
    ],
    highlight: true,
  },
  {
    nameKey: "pricing.teamsName",
    price: "$49",
    periodKey: "pricing.userMonth",
    descKey: "pricing.teamsDesc",
    ctaKey: "pricing.teamsCta",
    href: ROUTES.contact,
    featureKeys: [
      "pricing.f_everythingPro",
      "pricing.f_teamDash",
      "pricing.f_analytics",
      "pricing.f_csm",
      "pricing.f_sso",
    ],
    highlight: false,
  },
];

const faqs = [
  { qKey: "pricing.faq1Q", aKey: "pricing.faq1A" },
  { qKey: "pricing.faq2Q", aKey: "pricing.faq2A" },
  { qKey: "pricing.faq3Q", aKey: "pricing.faq3A" },
];

export default function PricingPage() {
  const t = useT();
  return (
    <div className="mx-auto max-w-7xl px-6 py-16">
      <div className="mx-auto max-w-2xl text-center">
        <h1 className="text-4xl font-extrabold tracking-tight">{t("pricing.title")}</h1>
        <p className="mt-4 text-lg text-muted-foreground">{t("pricing.subtitle")}</p>
      </div>

      <div className="mt-14 grid gap-6 lg:grid-cols-3">
        {plans.map((plan) => (
          <Card
            key={plan.nameKey}
            className={cn("flex flex-col p-8", plan.highlight && "border-primary shadow-lg ring-1 ring-primary")}
          >
            <div className="flex items-center justify-between">
              <h2 className="text-xl font-bold">{t(plan.nameKey)}</h2>
              {plan.highlight && <Badge>{t("pricing.mostPopular")}</Badge>}
            </div>
            <p className="mt-2 text-sm text-muted-foreground">{t(plan.descKey)}</p>
            <div className="mt-6 flex items-end gap-1">
              <span className="text-4xl font-extrabold">{plan.price}</span>
              <span className="pb-1 text-sm text-muted-foreground">{t(plan.periodKey)}</span>
            </div>
            <Button asChild className="mt-6" size="lg" variant={plan.highlight ? "default" : "outline"}>
              <Link href={plan.href}>{t(plan.ctaKey)}</Link>
            </Button>
            <ul className="mt-8 space-y-3 text-sm">
              {plan.featureKeys.map((f) => (
                <li key={f} className="flex items-start gap-2">
                  <Check className="mt-0.5 size-4 shrink-0 text-emerald-600" />
                  <span>{t(f)}</span>
                </li>
              ))}
            </ul>
          </Card>
        ))}
      </div>

      <section className="mx-auto mt-20 max-w-3xl">
        <h2 className="text-center text-2xl font-extrabold">{t("pricing.faqTitle")}</h2>
        <div className="mt-8 space-y-3">
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
    </div>
  );
}
