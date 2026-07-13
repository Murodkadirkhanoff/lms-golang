"use client";

import { useState } from "react";
import { Mail, MapPin, MessageSquare, Phone } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { FormField } from "@/components/ui/form-field";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { useT } from "@/providers/locale-provider";

export default function ContactPage() {
  const t = useT();
  const [sent, setSent] = useState(false);

  const channels = [
    { icon: Mail, label: t("contact.email"), value: "support@learnhub.com" },
    { icon: Phone, label: t("contact.phone"), value: "+998 71 200 00 00" },
    { icon: MapPin, label: t("contact.office"), value: t("contact.officeValue") },
  ];

  return (
    <div className="mx-auto max-w-5xl px-6 py-16">
      <div className="max-w-2xl">
        <h1 className="text-4xl font-extrabold tracking-tight">{t("contact.title")}</h1>
        <p className="mt-3 text-lg text-muted-foreground">{t("contact.subtitle")}</p>
      </div>

      <div className="mt-12 grid gap-8 lg:grid-cols-3">
        <div className="space-y-4 lg:col-span-1">
          {channels.map((c) => (
            <Card key={c.label} className="flex items-center gap-4 p-5">
              <div className="grid size-11 place-items-center rounded-xl bg-accent text-accent-foreground">
                <c.icon className="size-5" />
              </div>
              <div>
                <p className="text-sm font-semibold">{c.label}</p>
                <p className="text-sm text-muted-foreground">{c.value}</p>
              </div>
            </Card>
          ))}
        </div>

        <Card className="p-6 lg:col-span-2">
          {sent ? (
            <div className="flex flex-col items-center gap-3 py-12 text-center">
              <div className="grid size-12 place-items-center rounded-full bg-emerald-100 text-emerald-600">
                <MessageSquare className="size-6" />
              </div>
              <h2 className="text-lg font-bold">{t("contact.sentTitle")}</h2>
              <p className="text-sm text-muted-foreground">{t("contact.sentDesc")}</p>
              <Button variant="outline" onClick={() => setSent(false)}>
                {t("contact.sendAnother")}
              </Button>
            </div>
          ) : (
            <form
              onSubmit={(e) => {
                e.preventDefault();
                setSent(true);
              }}
              className="space-y-4"
            >
              <div className="grid gap-4 sm:grid-cols-2">
                <FormField label={t("contact.name")} name="name" required />
                <FormField label={t("contact.email")} name="email" type="email" required />
              </div>
              <FormField label={t("contact.subject")} name="subject" required />
              <div className="space-y-1.5">
                <Label htmlFor="message">{t("contact.message")}</Label>
                <Textarea id="message" rows={6} placeholder={t("contact.messagePlaceholder")} required />
              </div>
              <Button type="submit" size="lg">{t("contact.send")}</Button>
            </form>
          )}
        </Card>
      </div>
    </div>
  );
}
