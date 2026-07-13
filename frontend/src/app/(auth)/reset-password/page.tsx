"use client";

import { useState } from "react";
import Link from "next/link";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { CheckCircle2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { authService } from "@/services/auth.service";
import { resetPasswordSchema, type ResetPasswordValues } from "@/features/auth/schemas";
import { ROUTES } from "@/constants";
import { useT } from "@/providers/locale-provider";

export default function ResetPasswordPage() {
  const t = useT();
  const [done, setDone] = useState(false);
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ResetPasswordValues>({ resolver: zodResolver(resetPasswordSchema) });

  async function onSubmit(values: ResetPasswordValues) {
    await authService.resetPassword(values.password);
    setDone(true);
  }

  if (done) {
    return (
      <div className="text-center">
        <div className="mx-auto grid size-12 place-items-center rounded-full bg-emerald-100 text-emerald-600">
          <CheckCircle2 className="size-6" />
        </div>
        <h1 className="mt-4 text-2xl font-extrabold">{t("auth.passUpdated")}</h1>
        <p className="mt-1 text-sm text-muted-foreground">{t("auth.passUpdatedDesc")}</p>
        <Button asChild className="mt-6">
          <Link href={ROUTES.login}>{t("auth.continueLogin")}</Link>
        </Button>
      </div>
    );
  }

  return (
    <div>
      <h1 className="text-2xl font-extrabold">{t("auth.newPassTitle")}</h1>
      <p className="mt-1 text-sm text-muted-foreground">{t("auth.newPassSubtitle")}</p>

      <form onSubmit={handleSubmit(onSubmit)} className="mt-6 space-y-4">
        <FormField
          label={t("auth.newPassword")}
          type="password"
          placeholder="••••••••"
          error={errors.password?.message}
          {...register("password")}
        />
        <FormField
          label={t("auth.confirmPassword")}
          type="password"
          placeholder="••••••••"
          error={errors.confirmPassword?.message}
          {...register("confirmPassword")}
        />
        <Button type="submit" className="w-full" size="lg" disabled={isSubmitting}>
          {isSubmitting ? t("auth.saving") : t("auth.resetPassword")}
        </Button>
      </form>

      <p className="mt-6 text-center text-sm text-muted-foreground">
        {t("auth.backTo")}{" "}
        <Link href={ROUTES.login} className="font-semibold text-primary hover:underline">
          {t("common.login")}
        </Link>
      </p>
    </div>
  );
}
