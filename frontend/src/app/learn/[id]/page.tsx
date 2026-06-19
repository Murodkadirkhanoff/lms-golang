"use client";

import { use, useMemo, useState } from "react";
import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import {
  ArrowLeft,
  CheckCircle2,
  ChevronDown,
  Circle,
  Download,
  FileText,
  PlayCircle,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Progress } from "@/components/ui/progress";
import { LoadingState } from "@/components/shared/states";
import { coursesService } from "@/services/courses.service";
import { ROUTES } from "@/constants";
import { cn } from "@/lib/utils";

export default function LearnPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const { data: course, isLoading } = useQuery({
    queryKey: ["learn", id],
    queryFn: () => coursesService.getById(id),
  });

  const allLessons = useMemo(() => (course?.modules ?? []).flatMap((m) => m.lessons), [course]);
  const [activeLessonId, setActiveLessonId] = useState<number | null>(null);
  const [completed, setCompleted] = useState<Set<number>>(new Set());

  if (isLoading || !course) return <LoadingState className="min-h-screen" />;

  const activeLesson = allLessons.find((l) => l.id === activeLessonId) ?? allLessons[0];
  const baseCompleted = allLessons.filter((l) => l.completed).length;
  const totalDone = new Set([...completed, ...allLessons.filter((l) => l.completed).map((l) => l.id)]).size;
  const progress = allLessons.length ? Math.round((totalDone / allLessons.length) * 100) : 0;

  const isDone = (lessonId: number) =>
    completed.has(lessonId) || allLessons.find((l) => l.id === lessonId)?.completed;

  return (
    <div className="min-h-screen bg-secondary/30">
      {/* Top bar */}
      <header className="sticky top-0 z-40 flex h-14 items-center justify-between bg-slate-900 px-4 text-white">
        <div className="flex min-w-0 items-center gap-3">
          <Link href={ROUTES.dashboard} className="text-slate-300 hover:text-white">
            <ArrowLeft className="size-5" />
          </Link>
          <span className="truncate text-sm font-semibold">{course.title}</span>
        </div>
        <div className="flex items-center gap-3">
          <div className="hidden items-center gap-2 text-xs text-slate-300 sm:flex">
            <div className="h-2 w-32 overflow-hidden rounded-full bg-slate-700">
              <div className="h-full bg-emerald-500" style={{ width: `${progress}%` }} />
            </div>
            {progress}% complete
          </div>
        </div>
      </header>

      <div className="flex flex-col lg:flex-row">
        {/* Main */}
        <main className="min-w-0 flex-1">
          {/* Video */}
          <div className="relative grid aspect-video w-full place-items-center bg-black text-white">
            <div className="text-center">
              <PlayCircle className="mx-auto size-16 opacity-80" />
              <p className="mt-2 text-sm text-slate-300">{activeLesson?.title}</p>
            </div>
          </div>

          {/* Lesson meta + tabs */}
          <div className="bg-background">
            <div className="flex items-start justify-between gap-4 px-6 pt-6">
              <div>
                <h1 className="text-xl font-bold">{activeLesson?.title}</h1>
                <p className="mt-1 text-sm text-muted-foreground">
                  Lesson {allLessons.findIndex((l) => l.id === activeLesson?.id) + 1} of {allLessons.length}
                </p>
              </div>
              <Button
                onClick={() => activeLesson && setCompleted((prev) => new Set(prev).add(activeLesson.id))}
                disabled={!!activeLesson && !!isDone(activeLesson.id)}
                className="shrink-0 bg-emerald-600 hover:bg-emerald-700"
              >
                <CheckCircle2 className="size-4" />
                {activeLesson && isDone(activeLesson.id) ? "Completed" : "Mark complete"}
              </Button>
            </div>

            <div className="px-6 py-6">
              <Tabs defaultValue="overview">
                <TabsList>
                  <TabsTrigger value="overview">Overview</TabsTrigger>
                  <TabsTrigger value="notes">Notes</TabsTrigger>
                  <TabsTrigger value="resources">Resources</TabsTrigger>
                  <TabsTrigger value="qa">Q&amp;A</TabsTrigger>
                </TabsList>

                <TabsContent value="overview">
                  <p className="text-sm leading-relaxed text-muted-foreground">{course.description}</p>
                </TabsContent>

                <TabsContent value="notes">
                  <Textarea placeholder="Add a note for this lesson…" rows={4} />
                  <Button className="mt-3" size="sm">
                    Save note
                  </Button>
                </TabsContent>

                <TabsContent value="resources">
                  <div className="grid gap-3 sm:grid-cols-2">
                    {["lesson-starter.zip", "cheatsheet.pdf"].map((f) => (
                      <a
                        key={f}
                        href="#"
                        className="flex items-center gap-3 rounded-xl border p-3 hover:bg-secondary/50"
                      >
                        <FileText className="size-5 text-primary" />
                        <span className="flex-1 truncate text-sm font-semibold">{f}</span>
                        <Download className="size-4 text-muted-foreground" />
                      </a>
                    ))}
                  </div>
                </TabsContent>

                <TabsContent value="qa">
                  <Textarea placeholder="Ask a question about this lesson…" rows={3} />
                  <Button className="mt-3" size="sm">
                    Post question
                  </Button>
                </TabsContent>
              </Tabs>
            </div>
          </div>
        </main>

        {/* Curriculum sidebar */}
        <aside className="w-full border-t bg-background lg:h-[calc(100vh-3.5rem)] lg:w-96 lg:overflow-y-auto lg:border-l lg:border-t-0">
          <div className="border-b p-4">
            <h2 className="font-bold">Course content</h2>
            <p className="text-xs text-muted-foreground">
              {totalDone} of {allLessons.length} lessons · {progress}% complete
            </p>
            <Progress value={progress} className="mt-2" />
          </div>
          {(course.modules ?? []).map((module) => (
            <Module key={module.id} title={module.title}>
              {module.lessons.map((lesson) => {
                const active = lesson.id === activeLesson?.id;
                const done = isDone(lesson.id);
                return (
                  <button
                    key={lesson.id}
                    onClick={() => setActiveLessonId(lesson.id)}
                    className={cn(
                      "flex w-full items-center gap-3 px-4 py-2.5 text-left text-sm",
                      active ? "border-l-2 border-primary bg-accent font-semibold text-accent-foreground" : "hover:bg-secondary/50",
                    )}
                  >
                    {done ? (
                      <CheckCircle2 className="size-4 text-emerald-500" />
                    ) : active ? (
                      <PlayCircle className="size-4 text-primary" />
                    ) : (
                      <Circle className="size-4 text-slate-300" />
                    )}
                    <span className="flex-1">{lesson.title}</span>
                  </button>
                );
              })}
            </Module>
          ))}
          {baseCompleted >= 0 && (
            <div className="p-4">
              <Link
                href={ROUTES.quiz(course.id)}
                className="flex items-center gap-3 rounded-xl bg-amber-50 p-3 text-sm font-semibold text-amber-800 ring-1 ring-amber-200"
              >
                📝 Take the section quiz →
              </Link>
            </div>
          )}
        </aside>
      </div>
    </div>
  );
}

function Module({ title, children }: { title: string; children: React.ReactNode }) {
  const [open, setOpen] = useState(true);
  return (
    <div className="border-b">
      <button
        onClick={() => setOpen((o) => !o)}
        className="flex w-full items-center justify-between bg-secondary/50 px-4 py-3 text-left text-sm font-semibold"
      >
        <span>{title}</span>
        <ChevronDown className={cn("size-4 transition-transform", open && "rotate-180")} />
      </button>
      {open && <div>{children}</div>}
    </div>
  );
}
