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

export default function LoginPage() {
  const router = useRouter();
  const [serverError, setServerError] = useState("");
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginValues>({ resolver: zodResolver(loginSchema) });

  async function onSubmit(values: LoginValues) {
    setServerError("");
    try {
      await authService.login(values);
      router.push(ROUTES.dashboard);
    } catch (err) {
      setServerError(err instanceof Error ? err.message : "Login failed");
    }
  }

  return (
    <div>
      <h1 className="text-2xl font-extrabold">Welcome back</h1>
      <p className="mt-1 text-sm text-muted-foreground">Log in to continue learning.</p>

      <form onSubmit={handleSubmit(onSubmit)} className="mt-6 space-y-4">
        {serverError && (
          <div className="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700">{serverError}</div>
        )}
        <FormField label="Email" type="email" placeholder="you@example.com" error={errors.email?.message} {...register("email")} />
        <FormField label="Password" type="password" placeholder="••••••••" error={errors.password?.message} {...register("password")} />
        <div className="flex justify-end">
          <Link href={ROUTES.forgotPassword} className="text-sm font-medium text-primary hover:underline">
            Forgot password?
          </Link>
        </div>
        <Button type="submit" className="w-full" size="lg" disabled={isSubmitting}>
          {isSubmitting ? "Logging in…" : "Log in"}
        </Button>
      </form>

      <p className="mt-6 text-center text-sm text-muted-foreground">
        Don&apos;t have an account?{" "}
        <Link href={ROUTES.register} className="font-semibold text-primary hover:underline">
          Sign up
        </Link>
      </p>
    </div>
  );
}
