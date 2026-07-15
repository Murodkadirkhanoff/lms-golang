"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { BookOpen, CreditCard, Lock, PlayCircle, ShieldCheck } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { FormField } from "@/components/ui/form-field";
import { LoadingState } from "@/components/shared/states";
import { useCart } from "@/features/cart/cart-context";
import { useCartLines } from "@/features/cart/use-cart-lines";
import { ordersService } from "@/services/orders.service";
import { useAuth } from "@/providers/auth-provider";
import { useT } from "@/providers/locale-provider";
import { ROUTES } from "@/constants";
import { formatPrice } from "@/lib/utils";

const PAYMENT_METHODS = [
  { id: "card", labelKey: "checkout.card", icon: CreditCard },
  { id: "paypal", labelKey: "checkout.paypal", icon: ShieldCheck },
] as const;

export default function CheckoutPage() {
  const router = useRouter();
  const cart = useCart();
  const { isAuthenticated } = useAuth();
  const queryClient = useQueryClient();
  const t = useT();
  const [method, setMethod] = useState<string>("card");
  const [serverError, setServerError] = useState("");

  const { lines, subtotal, isLoading } = useCartLines();

  // Soliq qo'shilmaydi — backend buyurtmani element narxlari yig'indisi
  // bilan saqlaydi; ko'rsatilgan summa saqlangan bilan bir xil bo'lishi shart.
  const total = subtotal;

  // Buyurtma backendda yaratiladi (POST /me/orders); to'lovning o'zi hozircha
  // simulyatsiya — server buyurtmani darhol "paid" qilib kirish ochadi.
  const checkoutMutation = useMutation({
    mutationFn: () =>
      ordersService.checkout(
        lines.map((l) =>
          l.kind === "course" ? { courseId: l.courseId } : { lessonId: l.lessonId },
        ),
        method,
      ),
    onSuccess: () => {
      cart.clear();
      queryClient.invalidateQueries({ queryKey: ["dashboard"] });
      queryClient.invalidateQueries({ queryKey: ["my-courses"] });
      router.push(ROUTES.checkoutSuccess);
    },
    onError: (err) => setServerError(err instanceof Error ? err.message : "Payment failed"),
  });
  const processing = checkoutMutation.isPending;

  const pay = (e: React.FormEvent) => {
    e.preventDefault();
    setServerError("");
    if (!isAuthenticated) {
      router.push(ROUTES.login);
      return;
    }
    checkoutMutation.mutate();
  };

  if (cart.count === 0) {
    return (
      <div className="mx-auto max-w-2xl px-6 py-20 text-center">
        <h1 className="text-2xl font-extrabold">{t("checkout.emptyTitle")}</h1>
        <p className="mt-2 text-muted-foreground">{t("checkout.emptyDesc")}</p>
        <Button asChild className="mt-6">
          <Link href={ROUTES.courses}>{t("home.browseCourses")}</Link>
        </Button>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-7xl px-6 py-10">
      <h1 className="text-3xl font-extrabold">{t("checkout.title")}</h1>

      {isLoading ? (
        <LoadingState className="min-h-[40vh]" />
      ) : (
        <form onSubmit={pay} className="mt-8 grid gap-8 lg:grid-cols-3">
          <div className="space-y-6 lg:col-span-2">
            {/* Billing */}
            <Card className="p-6">
              <h2 className="text-lg font-bold">{t("checkout.billing")}</h2>
              <div className="mt-4 grid gap-4 sm:grid-cols-2">
                <FormField label={t("checkout.fullName")} name="name" defaultValue="Amir Karimov" required />
                <FormField label={t("checkout.email")} name="email" type="email" defaultValue="amir@example.com" required />
                <FormField label={t("checkout.country")} name="country" defaultValue="Uzbekistan" required />
                <FormField label={t("checkout.zip")} name="zip" defaultValue="100000" required />
              </div>
            </Card>

            {/* Payment method */}
            <Card className="p-6">
              <h2 className="text-lg font-bold">{t("checkout.paymentMethod")}</h2>
              <div className="mt-4 space-y-3">
                {PAYMENT_METHODS.map((m) => (
                  <label
                    key={m.id}
                    className={`flex cursor-pointer items-center gap-3 rounded-xl border p-4 transition-colors ${
                      method === m.id ? "border-primary bg-accent/50" : "hover:bg-secondary/50"
                    }`}
                  >
                    <input
                      type="radio"
                      name="method"
                      value={m.id}
                      checked={method === m.id}
                      onChange={() => setMethod(m.id)}
                      className="accent-primary"
                    />
                    <m.icon className="size-5 text-muted-foreground" />
                    <span className="font-medium">{t(m.labelKey)}</span>
                  </label>
                ))}
              </div>

              {method === "card" && (
                <div className="mt-5 grid gap-4 sm:grid-cols-2">
                  <div className="sm:col-span-2">
                    <FormField label={t("checkout.cardNumber")} name="card" placeholder="4242 4242 4242 4242" required />
                  </div>
                  <FormField label={t("checkout.expiry")} name="expiry" placeholder="MM / YY" required />
                  <FormField label={t("checkout.cvc")} name="cvc" placeholder="123" required />
                </div>
              )}
            </Card>
          </div>

          {/* Order summary */}
          <aside>
            <div className="lg:sticky lg:top-24">
              <Card className="p-6">
                <h2 className="text-lg font-bold">{t("checkout.orderSummary")}</h2>
                <div className="mt-4 space-y-3">
                  {lines.map((line) => (
                    <div key={line.key} className="flex items-center gap-3">
                      <div className={`grid h-12 w-16 shrink-0 place-items-center rounded-md bg-gradient-to-br ${line.thumbnailColor} text-white`}>
                        {line.kind === "course" ? <BookOpen className="size-4" /> : <PlayCircle className="size-4" />}
                      </div>
                      <div className="min-w-0 flex-1">
                        <p className="line-clamp-2 text-sm font-medium">{line.title}</p>
                        <p className="text-xs text-muted-foreground">
                          {line.kind === "course" ? t("cart.fullCourse") : t("cart.singleLesson")}
                        </p>
                      </div>
                      <span className="text-sm font-semibold">{formatPrice(line.price)}</span>
                    </div>
                  ))}
                </div>
                <div className="mt-5 space-y-2 border-t pt-5 text-sm">
                  <Row label={t("checkout.subtotal")} value={formatPrice(subtotal)} />
                  <div className="flex items-center justify-between border-t pt-3 text-base font-extrabold">
                    <span>{t("cart.total")}</span>
                    <span>{formatPrice(total)}</span>
                  </div>
                </div>
                {serverError && (
                  <p className="mt-4 rounded-lg bg-rose-50 px-3 py-2 text-xs text-rose-700">{serverError}</p>
                )}
                <Button type="submit" className="mt-5 w-full" size="lg" disabled={processing}>
                  {processing ? t("checkout.processing") : t("checkout.pay", { x: formatPrice(total) })}
                </Button>
                <p className="mt-3 flex items-center justify-center gap-1 text-center text-xs text-muted-foreground">
                  <Lock className="size-3" /> {t("checkout.secure")}
                </p>
              </Card>
            </div>
          </aside>
        </form>
      )}
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between text-muted-foreground">
      <span>{label}</span>
      <span className="font-medium text-foreground">{value}</span>
    </div>
  );
}
