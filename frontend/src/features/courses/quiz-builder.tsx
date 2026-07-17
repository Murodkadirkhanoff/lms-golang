"use client";

import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { quizService, type UpsertQuizInput } from "@/services/quiz.service";
import { useT } from "@/providers/locale-provider";
import { useToast } from "@/providers/toast-provider";

interface DraftQuestion {
  question: string;
  options: string[];
  correctIndex: number;
}

const emptyQuestion = (): DraftQuestion => ({ question: "", options: ["", ""], correctIndex: 0 });

/** Studio: kurs uchun yakuniy test yaratish/tahrirlash (har kursga bitta quiz). */
export function QuizBuilder({ courseId }: { courseId: number }) {
  const t = useT();
  const toast = useToast();
  const queryClient = useQueryClient();

  const { data: existing, isLoading } = useQuery({
    queryKey: ["quiz", courseId],
    queryFn: () => quizService.getForCourse(courseId),
  });

  const [title, setTitle] = useState("");
  const [passingScore, setPassingScore] = useState(70);
  const [timeLimit, setTimeLimit] = useState(10);
  const [questions, setQuestions] = useState<DraftQuestion[]>([]);
  const [error, setError] = useState("");

  // Mavjud quiz yuklangach forma to'ldiriladi.
  useEffect(() => {
    if (!existing) return;
    setTitle(existing.title);
    setPassingScore(existing.passingScore);
    setTimeLimit(existing.timeLimitMinutes);
    setQuestions(
      existing.questions.map((q) => ({
        question: q.question,
        options: [...q.options],
        correctIndex: q.correctIndex,
      })),
    );
  }, [existing]);

  const mutation = useMutation({
    mutationFn: (input: UpsertQuizInput) => quizService.upsert(courseId, input),
    onSuccess: () => {
      toast.success(t("quizB.saved"));
      queryClient.invalidateQueries({ queryKey: ["quiz", courseId] });
    },
    onError: (err) => setError(err instanceof Error ? err.message : t("common.somethingWrong")),
  });

  const patchQuestion = (index: number, patch: Partial<DraftQuestion>) =>
    setQuestions((qs) => qs.map((q, i) => (i === index ? { ...q, ...patch } : q)));

  const save = () => {
    setError("");
    mutation.mutate({
      title,
      passingScore,
      timeLimitMinutes: timeLimit,
      questions: questions.map((q) => ({
        question: q.question,
        options: q.options.filter((o) => o.trim() !== ""),
        correctIndex: q.correctIndex,
      })),
    });
  };

  if (isLoading) return null;

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("quizB.title")}</CardTitle>
        <CardDescription>{t("quizB.desc")}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-5">
        <div className="grid gap-4 sm:grid-cols-3">
          <div className="space-y-1.5 sm:col-span-3">
            <Label htmlFor="quiz-title">{t("quizB.quizTitle")}</Label>
            <Input id="quiz-title" value={title} onChange={(e) => setTitle(e.target.value)}
                   placeholder={t("quizB.titlePlaceholder")} />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="quiz-pass">{t("quizB.passingScore")}</Label>
            <Input id="quiz-pass" type="number" min={0} max={100} value={passingScore}
                   onChange={(e) => setPassingScore(Number(e.target.value))} />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="quiz-time">{t("quizB.timeLimit")}</Label>
            <Input id="quiz-time" type="number" min={1} value={timeLimit}
                   onChange={(e) => setTimeLimit(Number(e.target.value))} />
          </div>
        </div>

        {questions.map((q, qi) => (
          <div key={qi} className="rounded-xl border p-4">
            <div className="flex items-start justify-between gap-3">
              <div className="flex-1 space-y-1.5">
                <Label>{t("quizB.questionN", { n: qi + 1 })}</Label>
                <Input value={q.question} onChange={(e) => patchQuestion(qi, { question: e.target.value })}
                       placeholder={t("quizB.questionPlaceholder")} />
              </div>
              <Button type="button" variant="ghost" size="sm" className="mt-6 text-rose-600"
                      onClick={() => setQuestions((qs) => qs.filter((_, i) => i !== qi))}>
                <Trash2 className="size-4" />
              </Button>
            </div>

            <div className="mt-3 space-y-2">
              {q.options.map((opt, oi) => (
                <div key={oi} className="flex items-center gap-2">
                  {/* To'g'ri javob belgilash */}
                  <input
                    type="radio"
                    name={`correct-${qi}`}
                    aria-label={t("quizB.markCorrect")}
                    checked={q.correctIndex === oi}
                    onChange={() => patchQuestion(qi, { correctIndex: oi })}
                  />
                  <Input value={opt}
                         placeholder={t("quizB.optionN", { n: oi + 1 })}
                         onChange={(e) =>
                           patchQuestion(qi, { options: q.options.map((o, i) => (i === oi ? e.target.value : o)) })
                         } />
                  {q.options.length > 2 && (
                    <Button type="button" variant="ghost" size="sm"
                            onClick={() =>
                              patchQuestion(qi, {
                                options: q.options.filter((_, i) => i !== oi),
                                correctIndex: q.correctIndex >= oi && q.correctIndex > 0 ? q.correctIndex - 1 : q.correctIndex,
                              })
                            }>
                      <Trash2 className="size-3.5" />
                    </Button>
                  )}
                </div>
              ))}
              <Button type="button" variant="outline" size="sm"
                      onClick={() => patchQuestion(qi, { options: [...q.options, ""] })}>
                <Plus className="size-3.5" /> {t("quizB.addOption")}
              </Button>
            </div>
          </div>
        ))}

        <Button type="button" variant="outline" onClick={() => setQuestions((qs) => [...qs, emptyQuestion()])}>
          <Plus className="size-4" /> {t("quizB.addQuestion")}
        </Button>

        {error && <p className="text-sm text-destructive">{error}</p>}

        <div>
          <Button type="button" onClick={save} disabled={mutation.isPending}>
            {mutation.isPending ? t("cf.saving") : t("quizB.save")}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
