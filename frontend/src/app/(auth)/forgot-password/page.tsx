"use client";

import { useState } from "react";
import Link from "next/link";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { CheckCircle2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { authService } from "@/services/auth.service";
import { forgotPasswordSchema, type ForgotPasswordValues } from "@/features/auth/schemas";
import { ROUTES } from "@/constants";
import { useT } from "@/providers/locale-provider";

export default function ForgotPasswordPage() {
  const t = useT();
  const [sent, setSent] = useState(false);
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ForgotPasswordValues>({ resolver: zodResolver(forgotPasswordSchema) });

  async function onSubmit(values: ForgotPasswordValues) {
    await authService.forgotPassword(values.email);
    setSent(true);
  }

  if (sent) {
    return (
      <div className="text-center">
        <div className="mx-auto grid size-12 place-items-center rounded-full bg-emerald-100 text-emerald-600">
          <CheckCircle2 className="size-6" />
        </div>
        <h1 className="mt-4 text-2xl font-extrabold">{t("auth.checkEmail")}</h1>
        <p className="mt-1 text-sm text-muted-foreground">{t("auth.resetSent")}</p>
        <Button asChild variant="outline" className="mt-6">
          <Link href={ROUTES.login}>{t("auth.backToLogin")}</Link>
        </Button>
      </div>
    );
  }

  return (
    <div>
      <h1 className="text-2xl font-extrabold">{t("auth.resetTitle")}</h1>
      <p className="mt-1 text-sm text-muted-foreground">{t("auth.resetSubtitle")}</p>

      <form onSubmit={handleSubmit(onSubmit)} className="mt-6 space-y-4">
        <FormField label={t("auth.email")} type="email" placeholder="you@example.com" error={errors.email?.message} {...register("email")} />
        <Button type="submit" className="w-full" size="lg" disabled={isSubmitting}>
          {isSubmitting ? t("auth.sending") : t("auth.sendReset")}
        </Button>
      </form>

      <p className="mt-6 text-center text-sm text-muted-foreground">
        {t("auth.rememberedIt")}{" "}
        <Link href={ROUTES.login} className="font-semibold text-primary hover:underline">
          {t("common.login")}
        </Link>
      </p>
    </div>
  );
}
