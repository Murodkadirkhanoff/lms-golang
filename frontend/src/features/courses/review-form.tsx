"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Star } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Textarea } from "@/components/ui/textarea";
import { coursesService } from "@/services/courses.service";
import { useT } from "@/providers/locale-provider";
import { useToast } from "@/providers/toast-provider";
import { cn } from "@/lib/utils";

/**
 * Kursga yozilgan foydalanuvchi uchun baho + izoh formasi.
 * Backend bitta user uchun bitta sharh saqlaydi (qayta yuborilsa yangilanadi).
 */
export function ReviewForm({ courseId, slug }: { courseId: number; slug: string }) {
  const t = useT();
  const toast = useToast();
  const queryClient = useQueryClient();
  const [rating, setRating] = useState(0);
  const [hovered, setHovered] = useState(0);
  const [comment, setComment] = useState("");
  const [error, setError] = useState("");

  const mutation = useMutation({
    mutationFn: () => coursesService.submitReview(courseId, rating, comment.trim()),
    onSuccess: () => {
      toast.success(t("review.thanks"));
      setComment("");
      setRating(0);
      // Sharhlar ro'yxati kurs detali ichida keladi — qayta yuklaymiz.
      queryClient.invalidateQueries({ queryKey: ["course", slug] });
    },
    onError: (err) => setError(err instanceof Error ? err.message : t("common.somethingWrong")),
  });

  const submit = () => {
    setError("");
    if (rating < 1) {
      setError(t("review.ratingRequired"));
      return;
    }
    mutation.mutate();
  };

  return (
    <Card className="p-6">
      <h2 className="text-xl font-bold">{t("review.write")}</h2>
      <div className="mt-4 flex items-center gap-2">
        <span className="text-sm text-muted-foreground">{t("review.yourRating")}:</span>
        <div className="flex" onMouseLeave={() => setHovered(0)}>
          {[1, 2, 3, 4, 5].map((n) => (
            <button
              key={n}
              type="button"
              aria-label={`${n} / 5`}
              onClick={() => setRating(n)}
              onMouseEnter={() => setHovered(n)}
              className="p-0.5"
            >
              <Star
                className={cn(
                  "size-6 text-muted-foreground/40",
                  (hovered ? n <= hovered : n <= rating) && "fill-amber-400 text-amber-400",
                )}
              />
            </button>
          ))}
        </div>
      </div>
      <Textarea
        rows={3}
        value={comment}
        onChange={(e) => setComment(e.target.value)}
        placeholder={t("review.commentPlaceholder")}
        className="mt-3"
      />
      {error && <p className="mt-2 text-sm text-destructive">{error}</p>}
      <Button className="mt-3" onClick={submit} disabled={mutation.isPending}>
        {mutation.isPending ? t("review.submitting") : t("review.submit")}
      </Button>
    </Card>
  );
}
