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

export default function ForgotPasswordPage() {
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
        <h1 className="mt-4 text-2xl font-extrabold">Check your email</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          We&apos;ve sent a password reset link to your inbox.
        </p>
        <Button asChild variant="outline" className="mt-6">
          <Link href={ROUTES.login}>Back to login</Link>
        </Button>
      </div>
    );
  }

  return (
    <div>
      <h1 className="text-2xl font-extrabold">Reset your password</h1>
      <p className="mt-1 text-sm text-muted-foreground">
        Enter your email and we&apos;ll send you a reset link.
      </p>

      <form onSubmit={handleSubmit(onSubmit)} className="mt-6 space-y-4">
        <FormField label="Email" type="email" placeholder="you@example.com" error={errors.email?.message} {...register("email")} />
        <Button type="submit" className="w-full" size="lg" disabled={isSubmitting}>
          {isSubmitting ? "Sending…" : "Send reset link"}
        </Button>
      </form>

      <p className="mt-6 text-center text-sm text-muted-foreground">
        Remembered it?{" "}
        <Link href={ROUTES.login} className="font-semibold text-primary hover:underline">
          Log in
        </Link>
      </p>
    </div>
  );
}
