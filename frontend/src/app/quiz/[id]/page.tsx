"use client";

import { use, useEffect, useState } from "react";
import Link from "next/link";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Clock, CheckCircle2, XCircle, RotateCcw, ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { LoadingState } from "@/components/shared/states";
import { quizService } from "@/services/quiz.service";
import { ROUTES } from "@/constants";
import { cn } from "@/lib/utils";
import { useAuth } from "@/providers/auth-provider";
import { useT } from "@/providers/locale-provider";

function attemptDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en", { month: "short", day: "numeric", year: "numeric" });
}

export default function QuizPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const t = useT();
  const { isAuthenticated } = useAuth();
  const queryClient = useQueryClient();
  const { data: quiz, isLoading } = useQuery({ queryKey: ["quiz", id], queryFn: () => quizService.getById(id) });

  // Oldingi urinishlar tarixi (score history) — backenddan.
  const { data: history } = useQuery({
    queryKey: ["quiz-attempts", id],
    queryFn: () => quizService.listAttempts(id),
    enabled: isAuthenticated,
  });

  const submitMutation = useMutation({
    mutationFn: (score: number) => quizService.submitAttempt(id, score),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["quiz-attempts", id] }),
  });

  const [started, setStarted] = useState(false);
  const [current, setCurrent] = useState(0);
  const [answers, setAnswers] = useState<Record<number, number>>({});
  const [submitted, setSubmitted] = useState(false);
  const [timeLeft, setTimeLeft] = useState(0);

  // Yakunlash: natijani hisoblab, tarix uchun backendga yozadi
  // (Submit tugmasi ham, vaqt tugashi ham shu yo'ldan o'tadi).
  const finish = () => {
    if (submitted || !quiz) return;
    setSubmitted(true);
    const finalScore = Math.round(
      (quiz.questions.filter((q) => answers[q.id] === q.correctIndex).length / quiz.questions.length) * 100,
    );
    if (isAuthenticated) submitMutation.mutate(finalScore);
  };

  useEffect(() => {
    if (quiz && started) setTimeLeft(quiz.timeLimitMinutes * 60);
  }, [quiz, started]);

  useEffect(() => {
    if (!started || submitted) return;
    if (timeLeft <= 0) {
      finish();
      return;
    }
    const t = setInterval(() => setTimeLeft((s) => s - 1), 1000);
    return () => clearInterval(t);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [started, submitted, timeLeft]);

  if (isLoading || !quiz) return <LoadingState className="min-h-screen" />;

  const score = Math.round(
    (quiz.questions.filter((q) => answers[q.id] === q.correctIndex).length / quiz.questions.length) * 100,
  );
  const passed = score >= quiz.passingScore;
  const mins = Math.floor(timeLeft / 60);
  const secs = timeLeft % 60;

  const reset = () => {
    setStarted(false);
    setSubmitted(false);
    setCurrent(0);
    setAnswers({});
  };

  return (
    <div className="min-h-screen bg-secondary/30">
      <header className="flex h-14 items-center justify-between bg-slate-900 px-4 text-white">
        <Link href={ROUTES.dashboard} className="flex items-center gap-2 text-sm text-slate-300 hover:text-white">
          <ArrowLeft className="size-5" /> {t("quiz.exit")}
        </Link>
        {started && !submitted && (
          <div className="flex items-center gap-2 rounded-lg bg-white/10 px-3 py-1.5 text-sm font-semibold">
            <Clock className="size-4" /> {mins}:{secs.toString().padStart(2, "0")}
          </div>
        )}
      </header>

      <div className="mx-auto max-w-2xl px-6 py-10">
        {/* Intro */}
        {!started && (
          <Card>
            <CardHeader>
              <CardTitle className="text-2xl">{quiz.title}</CardTitle>
              <CardDescription>
                {t("quiz.meta", { n: quiz.questions.length, min: quiz.timeLimitMinutes, pass: quiz.passingScore })}
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <Button size="lg" onClick={() => setStarted(true)}>
                {t("quiz.start")}
              </Button>
              {(history ?? []).length > 0 && (
                <div>
                  <h3 className="mb-2 text-sm font-semibold">{t("quiz.scoreHistory")}</h3>
                  <div className="space-y-2">
                    {(history ?? []).map((h) => (
                      <div key={h.id} className="flex items-center justify-between rounded-lg bg-secondary/60 px-3 py-2 text-sm">
                        <span className="text-muted-foreground">{attemptDate(h.createdAt)}</span>
                        <span className={cn("font-semibold", h.score >= quiz.passingScore ? "text-emerald-600" : "text-rose-600")}>
                          {h.score}%
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        )}

        {/* Questions */}
        {started && !submitted && (
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between text-sm text-muted-foreground">
                <span>
                  {t("quiz.questionOf", { i: current + 1, n: quiz.questions.length })}
                </span>
              </div>
              <Progress value={((current + 1) / quiz.questions.length) * 100} className="mt-2" />
              <CardTitle className="pt-4 text-xl">{quiz.questions[current].question}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              {quiz.questions[current].options.map((opt, idx) => {
                const selected = answers[quiz.questions[current].id] === idx;
                return (
                  <button
                    key={idx}
                    onClick={() => setAnswers((a) => ({ ...a, [quiz.questions[current].id]: idx }))}
                    className={cn(
                      "flex w-full items-center gap-3 rounded-xl border p-4 text-left text-sm transition-colors",
                      selected ? "border-primary bg-accent" : "hover:bg-secondary/50",
                    )}
                  >
                    <span
                      className={cn(
                        "grid size-6 shrink-0 place-items-center rounded-full border text-xs font-semibold",
                        selected && "border-primary bg-primary text-primary-foreground",
                      )}
                    >
                      {String.fromCharCode(65 + idx)}
                    </span>
                    {opt}
                  </button>
                );
              })}

              <div className="flex justify-between pt-4">
                <Button variant="outline" disabled={current === 0} onClick={() => setCurrent((c) => c - 1)}>
                  {t("quiz.previous")}
                </Button>
                {current < quiz.questions.length - 1 ? (
                  <Button onClick={() => setCurrent((c) => c + 1)}>{t("quiz.next")}</Button>
                ) : (
                  <Button onClick={finish}>{t("quiz.submit")}</Button>
                )}
              </div>
            </CardContent>
          </Card>
        )}

        {/* Result */}
        {submitted && (
          <Card className="text-center">
            <CardContent className="space-y-4 p-8">
              <div
                className={cn(
                  "mx-auto grid size-16 place-items-center rounded-full",
                  passed ? "bg-emerald-100 text-emerald-600" : "bg-rose-100 text-rose-600",
                )}
              >
                {passed ? <CheckCircle2 className="size-8" /> : <XCircle className="size-8" />}
              </div>
              <h2 className="text-2xl font-extrabold">{passed ? t("quiz.congrats") : t("quiz.keepPracticing")}</h2>
              <p className="text-muted-foreground">
                {t("quiz.scored", { score, result: passed ? t("quiz.passed") : t("quiz.failed"), pass: quiz.passingScore })}
              </p>
              <div className="flex justify-center gap-3 pt-2">
                <Button variant="outline" onClick={reset}>
                  <RotateCcw className="size-4" /> {t("quiz.retake")}
                </Button>
                <Button asChild>
                  <Link href={ROUTES.dashboard}>{t("quiz.backToDashboard")}</Link>
                </Button>
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}
