"use client";

import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useT } from "@/providers/locale-provider";

export default function ProfilePage() {
  const t = useT();
  const learningStats = [
    { label: t("profile.statHours"), value: "142h" },
    { label: t("profile.statCourses"), value: "5" },
    { label: t("profile.statStreak"), value: t("profile.streakValue") },
    { label: t("profile.statQuizzes"), value: "23" },
  ];

  return (
    <div className="max-w-3xl space-y-6">
      <div className="flex items-center gap-4">
        <div className="size-16 rounded-full bg-indigo-200" />
        <div>
          <h1 className="text-2xl font-extrabold">Amir Karimov</h1>
          <p className="text-muted-foreground">amir@mail.com</p>
        </div>
      </div>

      <Tabs defaultValue="info">
        <TabsList>
          <TabsTrigger value="info">{t("profile.tabInfo")}</TabsTrigger>
          <TabsTrigger value="password">{t("profile.tabPassword")}</TabsTrigger>
          <TabsTrigger value="stats">{t("profile.tabStats")}</TabsTrigger>
        </TabsList>

        <TabsContent value="info">
          <Card>
            <CardHeader>
              <CardTitle>{t("profile.infoTitle")}</CardTitle>
              <CardDescription>{t("profile.infoDesc")}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 sm:grid-cols-2">
                <FormField label={t("profile.fullName")} defaultValue="Amir Karimov" />
                <FormField label={t("profile.email")} type="email" defaultValue="amir@mail.com" />
              </div>
              <FormField label={t("profile.headline")} defaultValue="Lifelong learner" />
              <Button>{t("common.saveChanges")}</Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="password">
          <Card>
            <CardHeader>
              <CardTitle>{t("profile.passTitle")}</CardTitle>
              <CardDescription>{t("profile.passDesc")}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <FormField label={t("profile.currentPass")} type="password" />
              <FormField label={t("profile.newPass")} type="password" />
              <FormField label={t("profile.confirmPass")} type="password" />
              <Button>{t("profile.updatePass")}</Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="stats">
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {learningStats.map((s) => (
              <Card key={s.label} className="p-5">
                <div className="text-3xl font-extrabold text-primary">{s.value}</div>
                <div className="mt-1 text-sm text-muted-foreground">{s.label}</div>
              </Card>
            ))}
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
