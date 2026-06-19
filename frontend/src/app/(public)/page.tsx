import Link from "next/link";
import {
  BarChart3,
  GraduationCap,
  MessageSquare,
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

const stats = [
  { value: "2,500+", label: "Courses" },
  { value: "320K", label: "Students" },
  { value: "850+", label: "Instructors" },
  { value: "98%", label: "Satisfaction" },
];

const features = [
  { icon: Target, title: "Structured Learning Paths", desc: "Guided curricula with modules, lessons and quizzes that adapt to your pace.", color: "bg-accent text-accent-foreground" },
  { icon: BarChart3, title: "Progress Tracking", desc: "Visual dashboards, streaks, and completion stats to keep you motivated.", color: "bg-emerald-100 text-emerald-600" },
  { icon: GraduationCap, title: "Verified Certificates", desc: "Earn shareable certificates recognized by employers worldwide.", color: "bg-amber-100 text-amber-600" },
  { icon: Users, title: "Expert Instructors", desc: "Learn from industry leaders with real-world experience.", color: "bg-rose-100 text-rose-600" },
  { icon: MessageSquare, title: "Community & Q&A", desc: "Discuss lessons, ask questions and get help from peers.", color: "bg-violet-100 text-violet-600" },
  { icon: Smartphone, title: "Learn Anywhere", desc: "Fully responsive — pick up where you left off on any device.", color: "bg-sky-100 text-sky-600" },
];

const testimonials = [
  { name: "Amir Karimov", role: "Frontend Developer", color: "bg-indigo-200", quote: "I switched careers into tech in 6 months. The structured paths and projects made all the difference." },
  { name: "Laura Bennett", role: "Product Designer", color: "bg-emerald-200", quote: "The quizzes and progress tracking kept me accountable. Best learning investment I've made." },
  { name: "David Park", role: "Instructor", color: "bg-amber-200", quote: "As an instructor I reached 12,000 students in a year. The dashboard analytics are fantastic." },
];

const faqs = [
  { q: "Do I get lifetime access to courses?", a: "Yes. Once you enroll, you have lifetime access to all course materials, including future updates." },
  { q: "Are certificates recognized?", a: "Certificates are verifiable and shareable on LinkedIn and with employers." },
  { q: "Can I get a refund?", a: "We offer a 30-day money-back guarantee, no questions asked." },
  { q: "Can I become an instructor?", a: "Absolutely. Anyone can open the Studio and publish a course — no separate role needed." },
];

export default function LandingPage() {
  return (
    <>
      {/* HERO */}
      <section className="relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-b from-accent/60 to-background" />
        <div className="relative mx-auto grid max-w-7xl items-center gap-12 px-6 py-20 lg:grid-cols-2">
          <div>
            <Badge>🚀 12,000+ learners this month</Badge>
            <h1 className="mt-5 text-5xl font-extrabold leading-tight tracking-tight">
              Learn without limits, <span className="text-primary">grow your career</span>
            </h1>
            <p className="mt-5 max-w-md text-lg text-muted-foreground">
              Build in-demand skills with expert-led courses. Learn at your own pace, track progress, and earn
              certificates.
            </p>
            <div className="mt-8 flex flex-col gap-3 sm:flex-row">
              <Button asChild size="lg">
                <Link href={ROUTES.courses}>Explore courses</Link>
              </Button>
              <Button asChild size="lg" variant="outline">
                <Link href={ROUTES.studio}>Start teaching</Link>
              </Button>
            </div>
            <div className="mt-8 flex items-center gap-4">
              <div className="flex -space-x-2">
                {["bg-indigo-200", "bg-emerald-200", "bg-amber-200", "bg-rose-200"].map((c) => (
                  <div key={c} className={`size-9 rounded-full ${c} ring-2 ring-background`} />
                ))}
              </div>
              <p className="text-sm text-muted-foreground">
                <span className="font-semibold text-foreground">4.8/5</span> from 9,200+ reviews
              </p>
            </div>
          </div>
          <div className="relative">
            <Card className="p-4 shadow-2xl">
              <div className="grid aspect-[4/3] place-items-center rounded-xl bg-gradient-to-br from-primary to-violet-600 text-primary-foreground">
                <div className="text-center">
                  <div className="mx-auto grid size-16 place-items-center rounded-full bg-white/20 text-2xl">▶</div>
                  <p className="mt-3 font-semibold">Course preview</p>
                </div>
              </div>
            </Card>
            <Card className="absolute -bottom-5 -left-5 p-4">
              <p className="text-xs text-muted-foreground">Course completed</p>
              <p className="font-bold text-emerald-600">+ Certificate earned 🎓</p>
            </Card>
          </div>
        </div>
      </section>

      {/* STATS */}
      <section className="border-y bg-secondary/40">
        <div className="mx-auto grid max-w-7xl grid-cols-2 gap-8 px-6 py-12 text-center md:grid-cols-4">
          {stats.map((s) => (
            <div key={s.label}>
              <div className="text-4xl font-extrabold text-primary">{s.value}</div>
              <div className="mt-1 text-sm text-muted-foreground">{s.label}</div>
            </div>
          ))}
        </div>
      </section>

      {/* FEATURES */}
      <section className="mx-auto max-w-7xl px-6 py-20">
        <div className="mx-auto max-w-2xl text-center">
          <h2 className="text-3xl font-extrabold">Everything you need to learn effectively</h2>
          <p className="mt-3 text-muted-foreground">A complete platform built for students, instructors and teams.</p>
        </div>
        <div className="mt-12 grid gap-6 md:grid-cols-3">
          {features.map((f) => (
            <Card key={f.title} className="p-6 transition-shadow hover:shadow-md">
              <div className={`grid size-12 place-items-center rounded-xl ${f.color}`}>
                <f.icon className="size-6" />
              </div>
              <h3 className="mt-4 text-lg font-bold">{f.title}</h3>
              <p className="mt-2 text-sm text-muted-foreground">{f.desc}</p>
            </Card>
          ))}
        </div>
      </section>

      {/* POPULAR COURSES */}
      <section className="border-y bg-secondary/40">
        <div className="mx-auto max-w-7xl px-6 py-20">
          <div className="flex items-end justify-between">
            <div>
              <h2 className="text-3xl font-extrabold">Popular courses</h2>
              <p className="mt-2 text-muted-foreground">Most enrolled this month</p>
            </div>
            <Button asChild variant="link">
              <Link href={ROUTES.courses}>View all →</Link>
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
          <h2 className="text-3xl font-extrabold">Loved by learners worldwide</h2>
          <p className="mt-3 text-muted-foreground">Real stories from our community.</p>
        </div>
        <div className="mt-12 grid gap-6 md:grid-cols-3">
          {testimonials.map((t) => (
            <Card key={t.name} className="p-6">
              <div className="flex gap-0.5 text-amber-400">
                {Array.from({ length: 5 }).map((_, i) => (
                  <Star key={i} className="size-4 fill-amber-400" />
                ))}
              </div>
              <blockquote className="mt-3 text-foreground/90">&ldquo;{t.quote}&rdquo;</blockquote>
              <div className="mt-5 flex items-center gap-3">
                <div className={`size-10 rounded-full ${t.color}`} />
                <div>
                  <div className="text-sm font-semibold">{t.name}</div>
                  <div className="text-xs text-muted-foreground">{t.role}</div>
                </div>
              </div>
            </Card>
          ))}
        </div>
      </section>

      {/* FAQ */}
      <section className="border-y bg-secondary/40">
        <div className="mx-auto max-w-3xl px-6 py-20">
          <h2 className="text-center text-3xl font-extrabold">Frequently asked questions</h2>
          <div className="mt-10 space-y-3">
            {faqs.map((f, i) => (
              <details key={f.q} className="group rounded-xl border bg-card p-5" open={i === 0}>
                <summary className="flex cursor-pointer items-center justify-between font-semibold">
                  {f.q}
                  <span className="text-primary transition-transform group-open:rotate-45">+</span>
                </summary>
                <p className="mt-3 text-sm text-muted-foreground">{f.a}</p>
              </details>
            ))}
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="mx-auto max-w-7xl px-6 py-20">
        <div className="rounded-3xl bg-gradient-to-br from-primary to-violet-700 px-8 py-16 text-center text-primary-foreground">
          <h2 className="text-3xl font-extrabold sm:text-4xl">Start learning today</h2>
          <p className="mx-auto mt-3 max-w-xl text-indigo-100">
            Join 320,000+ learners building their future. Your first course is on us.
          </p>
          <div className="mt-8 flex flex-col justify-center gap-3 sm:flex-row">
            <Button asChild size="lg" variant="secondary">
              <Link href={ROUTES.register}>Create free account</Link>
            </Button>
            <Button asChild size="lg" variant="outline" className="border-white/40 bg-transparent text-white hover:bg-white/10">
              <Link href={ROUTES.courses}>Browse courses</Link>
            </Button>
          </div>
        </div>
      </section>
    </>
  );
}
