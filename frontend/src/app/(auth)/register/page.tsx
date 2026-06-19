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

export default function RegisterPage() {
  const router = useRouter();
  const [serverError, setServerError] = useState("");
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<RegisterValues>({ resolver: zodResolver(registerSchema) });

  async function onSubmit(values: RegisterValues) {
    setServerError("");
    try {
      await authService.register(values);
      router.push(ROUTES.dashboard);
    } catch (err) {
      setServerError(err instanceof Error ? err.message : "Registration failed");
    }
  }

  return (
    <div>
      <h1 className="text-2xl font-extrabold">Create your account</h1>
      <p className="mt-1 text-sm text-muted-foreground">Start learning and teaching today — it&apos;s free.</p>

      <form onSubmit={handleSubmit(onSubmit)} className="mt-6 space-y-4">
        {serverError && (
          <div className="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700">{serverError}</div>
        )}
        <FormField label="Full name" placeholder="Jane Doe" error={errors.name?.message} {...register("name")} />
        <FormField label="Email" type="email" placeholder="you@example.com" error={errors.email?.message} {...register("email")} />
        <FormField label="Password" type="password" placeholder="At least 8 characters" error={errors.password?.message} {...register("password")} />
        <Button type="submit" className="w-full" size="lg" disabled={isSubmitting}>
          {isSubmitting ? "Creating account…" : "Create account"}
        </Button>
      </form>

      <p className="mt-6 text-center text-sm text-muted-foreground">
        Already have an account?{" "}
        <Link href={ROUTES.login} className="font-semibold text-primary hover:underline">
          Log in
        </Link>
      </p>
    </div>
  );
}
