"use client";

import Link from "next/link";
import {
  BarChart3,
  GraduationCap,
  MessageSquare,
  Play,
  Smartphone,
  Star,
  Target,
  Users,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { PopularCourses } from "@/features/courses/popular-courses";
import { ROUTES } from "@/constants";
import { useT } from "@/providers/locale-provider";

const stats = [
  { value: "2,500+", key: "home.statCourses" },
  { value: "320K", key: "home.statStudents" },
  { value: "850+", key: "home.statInstructors" },
  { value: "98%", key: "home.statSatisfaction" },
];

const features = [
  { icon: Target, titleKey: "home.feat1Title", descKey: "home.feat1Desc", color: "bg-accent text-accent-foreground" },
  { icon: BarChart3, titleKey: "home.feat2Title", descKey: "home.feat2Desc", color: "bg-emerald-100 text-emerald-600" },
  { icon: GraduationCap, titleKey: "home.feat3Title", descKey: "home.feat3Desc", color: "bg-amber-100 text-amber-600" },
  { icon: Users, titleKey: "home.feat4Title", descKey: "home.feat4Desc", color: "bg-rose-100 text-rose-600" },
  { icon: MessageSquare, titleKey: "home.feat5Title", descKey: "home.feat5Desc", color: "bg-violet-100 text-violet-600" },
  { icon: Smartphone, titleKey: "home.feat6Title", descKey: "home.feat6Desc", color: "bg-sky-100 text-sky-600" },
];

const testimonials = [
  { name: "Amir Karimov", roleKey: "home.t1Role", quoteKey: "home.t1Quote", color: "bg-indigo-200" },
  { name: "Laura Bennett", roleKey: "home.t2Role", quoteKey: "home.t2Quote", color: "bg-emerald-200" },
  { name: "David Park", roleKey: "home.t3Role", quoteKey: "home.t3Quote", color: "bg-amber-200" },
];

const faqs = [
  { qKey: "home.faq1Q", aKey: "home.faq1A" },
  { qKey: "home.faq2Q", aKey: "home.faq2A" },
  { qKey: "home.faq3Q", aKey: "home.faq3A" },
  { qKey: "home.faq4Q", aKey: "home.faq4A" },
];

export default function LandingPage() {
  const t = useT();
  return (
    <>
      {/* HERO */}
      <section className="relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-b from-accent/60 to-background" />
        <div className="relative mx-auto grid max-w-7xl items-center gap-12 px-6 py-20 lg:grid-cols-2">
          <div>
            <Badge>{t("home.badge")}</Badge>
            <h1 className="mt-5 text-5xl font-extrabold leading-tight tracking-tight">
              {t("home.heroTitle1")}<span className="text-primary">{t("home.heroTitle2")}</span>
            </h1>
            <p className="mt-5 max-w-md text-lg text-muted-foreground">{t("home.heroSubtitle")}</p>
            <div className="mt-8 flex flex-col gap-3 sm:flex-row">
              <Button asChild size="lg">
                <Link href={ROUTES.courses}>{t("home.exploreCourses")}</Link>
              </Button>
              <Button asChild size="lg" variant="outline">
                <Link href={ROUTES.studio}>{t("home.startTeaching")}</Link>
              </Button>
            </div>
            <div className="mt-8 flex items-center gap-4">
              <div className="flex -space-x-2">
                {["bg-indigo-200", "bg-emerald-200", "bg-amber-200", "bg-rose-200"].map((c) => (
                  <div key={c} className={`size-9 rounded-full ${c} ring-2 ring-background`} />
                ))}
              </div>
              <p className="text-sm text-muted-foreground">
                <span className="font-semibold text-foreground">4.8/5</span> {t("home.reviewsLine")}
              </p>
            </div>
          </div>
          <div className="relative">
            <Card className="p-4 shadow-2xl">
              <div className="grid aspect-[4/3] place-items-center rounded-xl bg-gradient-to-br from-primary to-violet-600 text-primary-foreground">
                <div className="text-center">
                  <div className="mx-auto grid size-16 place-items-center rounded-full bg-white/20">
                    <Play className="size-7 fill-current" aria-hidden="true" />
                  </div>
                  <p className="mt-3 font-semibold">{t("home.coursePreview")}</p>
                </div>
              </div>
            </Card>
            <Card className="absolute -bottom-5 -left-5 p-4">
              <p className="text-xs text-muted-foreground">{t("home.courseCompleted")}</p>
              <p className="font-bold text-emerald-600">{t("home.certEarned")}</p>
            </Card>
          </div>
        </div>
      </section>

      {/* STATS */}
      <section className="border-y bg-secondary/40">
        <div className="mx-auto grid max-w-7xl grid-cols-2 gap-8 px-6 py-12 text-center md:grid-cols-4">
          {stats.map((s) => (
            <div key={s.key}>
              <div className="text-4xl font-extrabold text-primary">{s.value}</div>
              <div className="mt-1 text-sm text-muted-foreground">{t(s.key)}</div>
            </div>
          ))}
        </div>
      </section>

      {/* FEATURES */}
      <section className="mx-auto max-w-7xl px-6 py-20">
        <div className="mx-auto max-w-2xl text-center">
          <h2 className="text-3xl font-extrabold">{t("home.featuresTitle")}</h2>
          <p className="mt-3 text-muted-foreground">{t("home.featuresSubtitle")}</p>
        </div>
        <div className="mt-12 grid gap-6 md:grid-cols-3">
          {features.map((f) => (
            <Card key={f.titleKey} className="p-6 transition-shadow hover:shadow-md">
              <div className={`grid size-12 place-items-center rounded-xl ${f.color}`}>
                <f.icon className="size-6" />
              </div>
              <h3 className="mt-4 text-lg font-bold">{t(f.titleKey)}</h3>
              <p className="mt-2 text-sm text-muted-foreground">{t(f.descKey)}</p>
            </Card>
          ))}
        </div>
      </section>

      {/* POPULAR COURSES */}
      <section className="border-y bg-secondary/40">
        <div className="mx-auto max-w-7xl px-6 py-20">
          <div className="flex items-end justify-between">
            <div>
              <h2 className="text-3xl font-extrabold">{t("home.popularTitle")}</h2>
              <p className="mt-2 text-muted-foreground">{t("home.popularSubtitle")}</p>
            </div>
            <Button asChild variant="link">
              <Link href={ROUTES.courses}>{t("common.viewAll")}</Link>
            </Button>
          </div>
          <div className="mt-10">
            <PopularCourses />
          </div>
        </div>
      </section>

      {/* TESTIMONIALS */}
      <section className="mx-auto max-w-7xl px-6 py-20">
        <div className="mx-auto max-w-2xl text-center">
          <h2 className="text-3xl font-extrabold">{t("home.testimonialsTitle")}</h2>
          <p className="mt-3 text-muted-foreground">{t("home.testimonialsSubtitle")}</p>
        </div>
        <div className="mt-12 grid gap-6 md:grid-cols-3">
          {testimonials.map((tm) => (
            <Card key={tm.name} className="p-6">
              <div className="flex gap-0.5 text-amber-400">
                {Array.from({ length: 5 }).map((_, i) => (
                  <Star key={i} className="size-4 fill-amber-400" />
                ))}
              </div>
              <blockquote className="mt-3 text-foreground/90">&ldquo;{t(tm.quoteKey)}&rdquo;</blockquote>
              <div className="mt-5 flex items-center gap-3">
                <div className={`size-10 rounded-full ${tm.color}`} />
                <div>
                  <div className="text-sm font-semibold">{tm.name}</div>
                  <div className="text-xs text-muted-foreground">{t(tm.roleKey)}</div>
                </div>
              </div>
            </Card>
          ))}
        </div>
      </section>

      {/* FAQ */}
      <section className="border-y bg-secondary/40">
        <div className="mx-auto max-w-3xl px-6 py-20">
          <h2 className="text-center text-3xl font-extrabold">{t("home.faqTitle")}</h2>
          <div className="mt-10 space-y-3">
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
        </div>
      </section>

      {/* CTA */}
      <section className="mx-auto max-w-7xl px-6 py-20">
        <div className="rounded-3xl bg-gradient-to-br from-primary to-violet-700 px-8 py-16 text-center text-primary-foreground">
          <h2 className="text-3xl font-extrabold sm:text-4xl">{t("home.ctaTitle")}</h2>
          <p className="mx-auto mt-3 max-w-xl text-indigo-100">{t("home.ctaSubtitle")}</p>
          <div className="mt-8 flex flex-col justify-center gap-3 sm:flex-row">
            <Button asChild size="lg" variant="secondary">
              <Link href={ROUTES.register}>{t("home.createFree")}</Link>
            </Button>
            <Button asChild size="lg" variant="outline" className="border-white/40 bg-transparent text-white hover:bg-white/10">
              <Link href={ROUTES.courses}>{t("home.browseCourses")}</Link>
            </Button>
          </div>
        </div>
      </section>
    </>
  );
}
