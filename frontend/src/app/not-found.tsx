"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";
import { ROUTES } from "@/constants";
import { useT } from "@/providers/locale-provider";

export default function NotFound() {
  const t = useT();
  return (
    <div className="flex min-h-screen flex-col items-center justify-center px-6 text-center">
      <p className="text-7xl font-extrabold text-primary">404</p>
      <h1 className="mt-4 text-2xl font-extrabold">{t("nf.title")}</h1>
      <p className="mt-2 max-w-md text-muted-foreground">{t("nf.desc")}</p>
      <div className="mt-8 flex gap-3">
        <Button asChild>
          <Link href={ROUTES.home}>{t("nf.home")}</Link>
        </Button>
        <Button asChild variant="outline">
          <Link href={ROUTES.courses}>{t("home.browseCourses")}</Link>
        </Button>
      </div>
    </div>
  );
}
