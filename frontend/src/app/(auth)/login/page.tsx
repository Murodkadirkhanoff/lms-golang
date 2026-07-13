"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { authService } from "@/services/auth.service";
import { loginSchema, type LoginValues } from "@/features/auth/schemas";
import { ROUTES } from "@/constants";
import { useT } from "@/providers/locale-provider";
import { useAuth } from "@/providers/auth-provider";

export default function LoginPage() {
  const router = useRouter();
  const t = useT();
  const { setUser } = useAuth();
  const [serverError, setServerError] = useState("");
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginValues>({ resolver: zodResolver(loginSchema) });

  async function onSubmit(values: LoginValues) {
    setServerError("");
    try {
      const { user } = await authService.login(values);
      setUser(user);
      router.push(ROUTES.dashboard);
    } catch (err) {
      setServerError(err instanceof Error ? err.message : t("auth.loginFailed"));
    }
  }

  return (
    <div>
      <h1 className="text-2xl font-extrabold">{t("auth.welcomeBack")}</h1>
      <p className="mt-1 text-sm text-muted-foreground">{t("auth.loginSubtitle")}</p>

      <form onSubmit={handleSubmit(onSubmit)} className="mt-6 space-y-4">
        {serverError && (
          <div role="alert" className="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700">
            {serverError}
          </div>
        )}
        <FormField label={t("auth.email")} type="email" placeholder="you@example.com" error={errors.email?.message} {...register("email")} />
        <FormField label={t("auth.password")} type="password" placeholder="••••••••" error={errors.password?.message} {...register("password")} />
        <div className="flex justify-end">
          <Link href={ROUTES.forgotPassword} className="text-sm font-medium text-primary hover:underline">
            {t("auth.forgotQ")}
          </Link>
        </div>
        <Button type="submit" className="w-full" size="lg" disabled={isSubmitting}>
          {isSubmitting ? t("auth.loggingIn") : t("common.login")}
        </Button>
      </form>

      <p className="mt-6 text-center text-sm text-muted-foreground">
        {t("auth.noAccount")}{" "}
        <Link href={ROUTES.register} className="font-semibold text-primary hover:underline">
          {t("auth.signUp")}
        </Link>
      </p>
    </div>
  );
}
