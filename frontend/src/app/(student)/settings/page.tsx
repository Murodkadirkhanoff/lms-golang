"use client";

import { CreditCard, Plus, Trash2 } from "lucide-react";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { LANGUAGES } from "@/constants";
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
              <div className="size-16 rounded-full bg-indigo-200" />
              <div>
                <Button variant="outline" size="sm">{t("settings.changePhoto")}</Button>
                <p className="mt-1 text-xs text-muted-foreground">{t("settings.photoHint")}</p>
              </div>
            </div>
            <div className="mt-6 grid gap-4 sm:grid-cols-2">
              <FormField label={t("profile.fullName")} name="name" defaultValue="Amir Karimov" />
              <FormField label={t("profile.email")} name="email" type="email" defaultValue="amir@example.com" />
              <div className="sm:col-span-2">
                <FormField label={t("profile.headline")} name="headline" defaultValue="Frontend Developer · Lifelong learner" />
              </div>
              <div className="sm:col-span-2 space-y-1.5">
                <Label htmlFor="bio">{t("settings.bio")}</Label>
                <Textarea id="bio" rows={4} defaultValue="Building delightful web experiences and always learning something new." />
              </div>
              <div className="space-y-1.5">
                <Label>{t("settings.prefLanguage")}</Label>
                <Select defaultValue="en">
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {LANGUAGES.map((l) => (
                      <SelectItem key={l.value} value={l.value}>
                        {l.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="mt-6 flex justify-end gap-2">
              <Button variant="outline">{t("common.cancel")}</Button>
              <Button>{t("common.saveChanges")}</Button>
            </div>
          </Card>
        </TabsContent>

        {/* Security */}
        <TabsContent value="security">
          <Card className="p-6">
            <h2 className="font-bold">{t("settings.changePassword")}</h2>
            <div className="mt-4 grid gap-4">
              <FormField label={t("profile.currentPass")} name="current" type="password" />
              <FormField label={t("profile.newPass")} name="new" type="password" />
              <FormField label={t("profile.confirmPass")} name="confirm" type="password" />
            </div>
            <div className="mt-4 flex justify-end">
              <Button>{t("profile.updatePass")}</Button>
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
