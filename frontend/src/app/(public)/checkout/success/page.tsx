"use client";

import Link from "next/link";
import { CheckCircle2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { ROUTES } from "@/constants";
import { useT } from "@/providers/locale-provider";

export default function CheckoutSuccessPage() {
  const t = useT();
  return (
    <div className="mx-auto flex min-h-[70vh] max-w-2xl flex-col items-center justify-center px-6 py-16 text-center">
      <div className="grid size-20 place-items-center rounded-full bg-emerald-100 text-emerald-600">
        <CheckCircle2 className="size-12" />
      </div>
      <h1 className="mt-6 text-3xl font-extrabold">{t("success.title")}</h1>
      <p className="mt-3 max-w-md text-muted-foreground">{t("success.body")}</p>
      <div className="mt-8 flex flex-col gap-3 sm:flex-row">
        <Button asChild size="lg">
          <Link href={ROUTES.myCourses}>{t("success.startLearning")}</Link>
        </Button>
        <Button asChild size="lg" variant="outline">
          <Link href={ROUTES.purchases}>{t("success.viewPurchases")}</Link>
        </Button>
      </div>
      <Card className="mt-10 w-full p-6 text-left">
        <h2 className="font-bold">{t("success.next")}</h2>
        <ul className="mt-3 space-y-2 text-sm text-muted-foreground">
          <li className="flex items-center gap-2"><CheckCircle2 className="size-4 text-emerald-600" /> {t("success.next1")}</li>
          <li className="flex items-center gap-2"><CheckCircle2 className="size-4 text-emerald-600" /> {t("success.next2")}</li>
          <li className="flex items-center gap-2"><CheckCircle2 className="size-4 text-emerald-600" /> {t("success.next3")}</li>
        </ul>
      </Card>
    </div>
  );
}
