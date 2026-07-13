"use client";

import Link from "next/link";
import { Heart, Lightbulb, Target } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { ROUTES } from "@/constants";
import { useT } from "@/providers/locale-provider";

const values = [
  { icon: Target, titleKey: "about.v1Title", descKey: "about.v1Desc" },
  { icon: Lightbulb, titleKey: "about.v2Title", descKey: "about.v2Desc" },
  { icon: Heart, titleKey: "about.v3Title", descKey: "about.v3Desc" },
];

const team = [
  { name: "Sarah Mitchell", roleKey: "about.role1", color: "bg-indigo-200" },
  { name: "James Carter", roleKey: "about.role2", color: "bg-emerald-200" },
  { name: "Elena Rodriguez", roleKey: "about.role3", color: "bg-amber-200" },
  { name: "Michael Chen", roleKey: "about.role4", color: "bg-rose-200" },
];

const stats = [
  { value: "320K+", key: "about.statLearners" },
  { value: "2,500+", key: "about.statCourses" },
  { value: "190", key: "about.statCountries" },
  { value: "98%", key: "about.statSatisfaction" },
];

export default function AboutPage() {
  const t = useT();
  return (
    <>
      <section className="border-b bg-secondary/40">
        <div className="mx-auto max-w-4xl px-6 py-20 text-center">
          <h1 className="text-4xl font-extrabold tracking-tight">{t("about.missionTitle")}</h1>
          <p className="mx-auto mt-5 max-w-2xl text-lg text-muted-foreground">{t("about.missionBody")}</p>
        </div>
      </section>

      <section className="mx-auto max-w-7xl px-6 py-16">
        <div className="grid grid-cols-2 gap-8 text-center md:grid-cols-4">
          {stats.map((s) => (
            <div key={s.key}>
              <div className="text-4xl font-extrabold text-primary">{s.value}</div>
              <div className="mt-1 text-sm text-muted-foreground">{t(s.key)}</div>
            </div>
          ))}
        </div>
      </section>

      <section className="border-y bg-secondary/40">
        <div className="mx-auto max-w-7xl px-6 py-20">
          <h2 className="text-center text-3xl font-extrabold">{t("about.valuesTitle")}</h2>
          <div className="mt-12 grid gap-6 md:grid-cols-3">
            {values.map((v) => (
              <Card key={v.titleKey} className="p-6">
                <div className="grid size-12 place-items-center rounded-xl bg-accent text-accent-foreground">
                  <v.icon className="size-6" />
                </div>
                <h3 className="mt-4 font-bold">{t(v.titleKey)}</h3>
                <p className="mt-2 text-sm text-muted-foreground">{t(v.descKey)}</p>
              </Card>
            ))}
          </div>
        </div>
      </section>

      <section className="mx-auto max-w-7xl px-6 py-20">
        <h2 className="text-center text-3xl font-extrabold">{t("about.teamTitle")}</h2>
        <div className="mt-12 grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
          {team.map((m) => (
            <Card key={m.name} className="p-6 text-center">
              <div className={`mx-auto size-20 rounded-full ${m.color}`} />
              <h3 className="mt-4 font-bold">{m.name}</h3>
              <p className="text-sm text-muted-foreground">{t(m.roleKey)}</p>
            </Card>
          ))}
        </div>
      </section>

      <section className="mx-auto max-w-7xl px-6 pb-20">
        <div className="rounded-3xl bg-gradient-to-br from-primary to-violet-700 px-8 py-16 text-center text-primary-foreground">
          <h2 className="text-3xl font-extrabold">{t("about.joinTitle")}</h2>
          <p className="mx-auto mt-3 max-w-xl text-indigo-100">{t("about.joinSubtitle")}</p>
          <div className="mt-8 flex justify-center gap-3">
            <Button asChild size="lg" variant="secondary">
              <Link href={ROUTES.courses}>{t("home.browseCourses")}</Link>
            </Button>
            <Button asChild size="lg" variant="outline" className="border-white/40 bg-transparent text-white hover:bg-white/10">
              <Link href={ROUTES.teach}>{t("about.becomeInstructor")}</Link>
            </Button>
          </div>
        </div>
      </section>
    </>
  );
}
