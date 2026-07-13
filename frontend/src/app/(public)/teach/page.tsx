"use client";

import Link from "next/link";
import { DollarSign, Globe2, Megaphone, Users, Video } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { ROUTES } from "@/constants";
import { useT } from "@/providers/locale-provider";

const benefits = [
  { icon: DollarSign, titleKey: "teach.b1Title", descKey: "teach.b1Desc" },
  { icon: Users, titleKey: "teach.b2Title", descKey: "teach.b2Desc" },
  { icon: Video, titleKey: "teach.b3Title", descKey: "teach.b3Desc" },
  { icon: Megaphone, titleKey: "teach.b4Title", descKey: "teach.b4Desc" },
];

const steps = [
  { n: 1, titleKey: "teach.s1Title", descKey: "teach.s1Desc" },
  { n: 2, titleKey: "teach.s2Title", descKey: "teach.s2Desc" },
  { n: 3, titleKey: "teach.s3Title", descKey: "teach.s3Desc" },
];

const stats = [
  { value: "850+", key: "teach.statInstructors" },
  { value: "$4.2M", key: "teach.statPaid" },
  { value: "320K", key: "teach.statReached" },
];

export default function TeachPage() {
  const t = useT();
  return (
    <>
      <section className="relative overflow-hidden border-b">
        <div className="absolute inset-0 bg-gradient-to-b from-accent/60 to-background" />
        <div className="relative mx-auto max-w-4xl px-6 py-20 text-center">
          <h1 className="text-4xl font-extrabold tracking-tight sm:text-5xl">
            {t("teach.heroTitle1")}<span className="text-primary">{t("teach.heroTitle2")}</span>
          </h1>
          <p className="mx-auto mt-5 max-w-xl text-lg text-muted-foreground">{t("teach.heroSubtitle")}</p>
          <div className="mt-8 flex justify-center gap-3">
            <Button asChild size="lg">
              <Link href={ROUTES.studioCourseNew}>{t("teach.startToday")}</Link>
            </Button>
            <Button asChild size="lg" variant="outline">
              <Link href={ROUTES.studio}>{t("teach.goToStudio")}</Link>
            </Button>
          </div>
        </div>
      </section>

      <section className="border-b bg-secondary/40">
        <div className="mx-auto grid max-w-5xl grid-cols-1 gap-8 px-6 py-12 text-center sm:grid-cols-3">
          {stats.map((s) => (
            <div key={s.key}>
              <div className="text-4xl font-extrabold text-primary">{s.value}</div>
              <div className="mt-1 text-sm text-muted-foreground">{t(s.key)}</div>
            </div>
          ))}
        </div>
      </section>

      <section className="mx-auto max-w-7xl px-6 py-20">
        <h2 className="text-center text-3xl font-extrabold">{t("teach.whyTitle")}</h2>
        <div className="mt-12 grid gap-6 md:grid-cols-2 lg:grid-cols-4">
          {benefits.map((b) => (
            <Card key={b.titleKey} className="p-6">
              <div className="grid size-12 place-items-center rounded-xl bg-accent text-accent-foreground">
                <b.icon className="size-6" />
              </div>
              <h3 className="mt-4 font-bold">{t(b.titleKey)}</h3>
              <p className="mt-2 text-sm text-muted-foreground">{t(b.descKey)}</p>
            </Card>
          ))}
        </div>
      </section>

      <section className="border-y bg-secondary/40">
        <div className="mx-auto max-w-5xl px-6 py-20">
          <h2 className="text-center text-3xl font-extrabold">{t("teach.howTitle")}</h2>
          <div className="mt-12 grid gap-8 md:grid-cols-3">
            {steps.map((s) => (
              <div key={s.n} className="text-center">
                <div className="mx-auto grid size-12 place-items-center rounded-full bg-primary text-lg font-bold text-primary-foreground">
                  {s.n}
                </div>
                <h3 className="mt-4 font-bold">{t(s.titleKey)}</h3>
                <p className="mt-2 text-sm text-muted-foreground">{t(s.descKey)}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="mx-auto max-w-7xl px-6 py-20">
        <div className="rounded-3xl bg-gradient-to-br from-primary to-violet-700 px-8 py-16 text-center text-primary-foreground">
          <Globe2 className="mx-auto size-10" />
          <h2 className="mt-4 text-3xl font-extrabold">{t("teach.ctaTitle")}</h2>
          <p className="mx-auto mt-3 max-w-xl text-indigo-100">{t("teach.ctaSubtitle")}</p>
          <Button asChild size="lg" variant="secondary" className="mt-8">
            <Link href={ROUTES.studioCourseNew}>{t("teach.createCourse")}</Link>
          </Button>
        </div>
      </section>
    </>
  );
}
