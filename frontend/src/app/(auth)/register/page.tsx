"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { authService } from "@/services/auth.service";
import { registerSchema, type RegisterValues } from "@/features/auth/schemas";
import { ROUTES } from "@/constants";
import { useT } from "@/providers/locale-provider";
import { useAuth } from "@/providers/auth-provider";

export default function RegisterPage() {
  const router = useRouter();
  const t = useT();
  const { setUser } = useAuth();
  const [serverError, setServerError] = useState("");
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<RegisterValues>({ resolver: zodResolver(registerSchema) });

  async function onSubmit(values: RegisterValues) {
    setServerError("");
    try {
      const { user } = await authService.register(values);
      setUser(user);
      router.push(ROUTES.dashboard);
    } catch (err) {
      setServerError(err instanceof Error ? err.message : t("auth.registerFailed"));
    }
  }

  return (
    <div>
      <h1 className="text-2xl font-extrabold">{t("auth.createTitle")}</h1>
      <p className="mt-1 text-sm text-muted-foreground">{t("auth.registerSubtitle")}</p>

      <form onSubmit={handleSubmit(onSubmit)} className="mt-6 space-y-4">
        {serverError && (
          <div role="alert" className="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700">
            {serverError}
          </div>
        )}
        <FormField label={t("auth.fullName")} placeholder="Jane Doe" error={errors.name?.message} {...register("name")} />
        <FormField label={t("auth.email")} type="email" placeholder="you@example.com" error={errors.email?.message} {...register("email")} />
        <FormField label={t("auth.password")} type="password" placeholder={t("auth.passwordMin")} error={errors.password?.message} {...register("password")} />
        <Button type="submit" className="w-full" size="lg" disabled={isSubmitting}>
          {isSubmitting ? t("auth.creating") : t("auth.createAccount")}
        </Button>
      </form>

      <p className="mt-6 text-center text-sm text-muted-foreground">
        {t("auth.haveAccount")}{" "}
        <Link href={ROUTES.login} className="font-semibold text-primary hover:underline">
          {t("common.login")}
        </Link>
      </p>
    </div>
  );
}
