"use client";

import { useEffect, useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { CreditCard, Plus, Trash2 } from "lucide-react";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { authService } from "@/services/auth.service";
import { useAuth, initials } from "@/providers/auth-provider";
import { useToast } from "@/providers/toast-provider";
import { useT } from "@/providers/locale-provider";

const NOTIFICATION_PREFS = [
  { id: "course", labelKey: "settings.notifCourse", descKey: "settings.notifCourseDesc", default: true },
  { id: "promo", labelKey: "settings.notifPromo", descKey: "settings.notifPromoDesc", default: false },
  { id: "messages", labelKey: "settings.notifMessages", descKey: "settings.notifMessagesDesc", default: true },
  { id: "digest", labelKey: "settings.notifDigest", descKey: "settings.notifDigestDesc", default: true },
];

const PAYMENT_METHODS = [
  { id: 1, brand: "Visa", last4: "4242", exp: "08/27", primary: true },
  { id: 2, brand: "Mastercard", last4: "5511", exp: "01/26", primary: false },
];

export default function SettingsPage() {
  const t = useT();
  const toast = useToast();
  const { user, setUser } = useAuth();

  const [name, setName] = useState("");
  const [currentPass, setCurrentPass] = useState("");
  const [newPass, setNewPass] = useState("");
  const [confirmPass, setConfirmPass] = useState("");
  const [passError, setPassError] = useState("");

  useEffect(() => {
    if (user) setName(user.name);
  }, [user]);

  const nameMutation = useMutation({
    mutationFn: () => authService.updateProfile(name.trim()),
    onSuccess: (updated) => {
      setUser(updated);
      toast.success(t("profile.saved"));
    },
    onError: (err) => toast.error(err instanceof Error ? err.message : t("common.somethingWrong")),
  });

  const passwordMutation = useMutation({
    mutationFn: () => authService.changePassword(currentPass, newPass),
    onSuccess: () => {
      toast.success(t("profile.passUpdated"));
      setCurrentPass("");
      setNewPass("");
      setConfirmPass("");
    },
    onError: (err) => setPassError(err instanceof Error ? err.message : t("common.somethingWrong")),
  });

  const submitPassword = () => {
    setPassError("");
    if (newPass !== confirmPass) {
      setPassError(t("profile.passMismatch"));
      return;
    }
    passwordMutation.mutate();
  };

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div>
        <h1 className="text-2xl font-extrabold">{t("settings.title")}</h1>
        <p className="text-muted-foreground">{t("settings.subtitle")}</p>
      </div>

      <Tabs defaultValue="account">
        <TabsList>
          <TabsTrigger value="account">{t("settings.tabAccount")}</TabsTrigger>
          <TabsTrigger value="security">{t("settings.tabSecurity")}</TabsTrigger>
          <TabsTrigger value="notifications">{t("settings.tabNotifications")}</TabsTrigger>
          <TabsTrigger value="billing">{t("settings.tabBilling")}</TabsTrigger>
        </TabsList>

        {/* Account */}
        <TabsContent value="account">
          <Card className="p-6">
            <div className="flex items-center gap-4">
              <div className="grid size-16 place-items-center rounded-full bg-indigo-200 text-lg font-bold text-indigo-700">
                {user ? initials(user.name) : ""}
              </div>
              <div>
                <p className="font-semibold">{user?.name}</p>
                <p className="text-xs text-muted-foreground">{user?.email}</p>
              </div>
            </div>
            <div className="mt-6 grid gap-4 sm:grid-cols-2">
              <FormField label={t("profile.fullName")} name="name" value={name}
                         onChange={(e) => setName(e.target.value)} />
              {/* Email — login identifikatori, o'zgartirilmaydi */}
              <FormField label={t("profile.email")} name="email" type="email" value={user?.email ?? ""} disabled />
            </div>
            <div className="mt-6 flex justify-end">
              <Button onClick={() => nameMutation.mutate()} disabled={nameMutation.isPending || !name.trim()}>
                {nameMutation.isPending ? t("cf.saving") : t("common.saveChanges")}
              </Button>
            </div>
          </Card>
        </TabsContent>

        {/* Security */}
        <TabsContent value="security">
          <Card className="p-6">
            <h2 className="font-bold">{t("settings.changePassword")}</h2>
            <div className="mt-4 grid gap-4">
              <FormField label={t("profile.currentPass")} name="current" type="password" value={currentPass}
                         onChange={(e) => setCurrentPass(e.target.value)} />
              <FormField label={t("profile.newPass")} name="new" type="password" value={newPass}
                         onChange={(e) => setNewPass(e.target.value)} />
              <FormField label={t("profile.confirmPass")} name="confirm" type="password" value={confirmPass}
                         onChange={(e) => setConfirmPass(e.target.value)} error={passError || undefined} />
            </div>
            <div className="mt-4 flex justify-end">
              <Button onClick={submitPassword} disabled={passwordMutation.isPending || !currentPass || !newPass}>
                {passwordMutation.isPending ? t("cf.saving") : t("profile.updatePass")}
              </Button>
            </div>
          </Card>

          <Card className="mt-6 flex items-center justify-between p-6">
            <div>
              <h2 className="font-bold">{t("settings.2faTitle")}</h2>
              <p className="mt-1 text-sm text-muted-foreground">{t("settings.2faDesc")}</p>
            </div>
            <Switch />
          </Card>

          <Card className="mt-6 flex items-center justify-between border-destructive/30 p-6">
            <div>
              <h2 className="font-bold text-destructive">{t("settings.deleteTitle")}</h2>
              <p className="mt-1 text-sm text-muted-foreground">{t("settings.deleteDesc")}</p>
            </div>
            <Button variant="destructive">{t("settings.delete")}</Button>
          </Card>
        </TabsContent>

        {/* Notifications */}
        <TabsContent value="notifications">
          <Card className="divide-y p-0">
            {NOTIFICATION_PREFS.map((p) => (
              <div key={p.id} className="flex items-center justify-between gap-4 px-6 py-4">
                <div>
                  <p className="font-semibold">{t(p.labelKey)}</p>
                  <p className="text-sm text-muted-foreground">{t(p.descKey)}</p>
                </div>
                <Switch defaultChecked={p.default} />
              </div>
            ))}
          </Card>
        </TabsContent>

        {/* Billing */}
        <TabsContent value="billing">
          <Card className="p-6">
            <div className="flex items-center justify-between">
              <h2 className="font-bold">{t("settings.paymentMethods")}</h2>
              <Button variant="outline" size="sm">
                <Plus className="size-4" /> {t("settings.addCard")}
              </Button>
            </div>
            <div className="mt-4 space-y-3">
              {PAYMENT_METHODS.map((m) => (
                <div key={m.id} className="flex items-center gap-4 rounded-xl border p-4">
                  <div className="grid size-10 place-items-center rounded-lg bg-secondary">
                    <CreditCard className="size-5 text-muted-foreground" />
                  </div>
                  <div className="flex-1">
                    <p className="font-semibold">
                      {m.brand} •••• {m.last4}
                      {m.primary && <span className="ml-2 rounded-full bg-accent px-2 py-0.5 text-xs text-accent-foreground">{t("settings.primary")}</span>}
                    </p>
                    <p className="text-xs text-muted-foreground">{t("settings.expires", { exp: m.exp })}</p>
                  </div>
                  <Button variant="ghost" size="icon" aria-label={t("settings.removeCard")}>
                    <Trash2 className="size-4" />
                  </Button>
                </div>
              ))}
            </div>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
