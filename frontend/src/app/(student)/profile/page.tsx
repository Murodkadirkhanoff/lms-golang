"use client";

import { useEffect, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { authService } from "@/services/auth.service";
import { dashboardService } from "@/services/dashboard.service";
import { useAuth, initials } from "@/providers/auth-provider";
import { useToast } from "@/providers/toast-provider";
import { useT } from "@/providers/locale-provider";

export default function ProfilePage() {
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

  // Real o'quv statistikasi — GET /me/stats.
  const { data: stats } = useQuery({ queryKey: ["dashboard", "stats"], queryFn: dashboardService.getStats });

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

  const learningStats = [
    { label: t("profile.statEnrolled"), value: stats?.enrolled ?? "—" },
    { label: t("profile.statInProgress"), value: stats?.inProgress ?? "—" },
    { label: t("profile.statCompleted"), value: stats?.completed ?? "—" },
    { label: t("profile.statCertificates"), value: stats?.certificates ?? "—" },
  ];

  return (
    <div className="max-w-3xl space-y-6">
      <div className="flex items-center gap-4">
        <div className="grid size-16 place-items-center rounded-full bg-indigo-200 text-lg font-bold text-indigo-700">
          {user ? initials(user.name) : ""}
        </div>
        <div>
          <h1 className="text-2xl font-extrabold">{user?.name}</h1>
          <p className="text-muted-foreground">{user?.email}</p>
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
                <FormField label={t("profile.fullName")} value={name} onChange={(e) => setName(e.target.value)} />
                {/* Email — login identifikatori, o'zgartirilmaydi */}
                <FormField label={t("profile.email")} type="email" value={user?.email ?? ""} disabled />
              </div>
              <Button onClick={() => nameMutation.mutate()} disabled={nameMutation.isPending || !name.trim()}>
                {nameMutation.isPending ? t("cf.saving") : t("common.saveChanges")}
              </Button>
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
              <FormField label={t("profile.currentPass")} type="password" value={currentPass}
                         onChange={(e) => setCurrentPass(e.target.value)} />
              <FormField label={t("profile.newPass")} type="password" value={newPass}
                         onChange={(e) => setNewPass(e.target.value)} />
              <FormField label={t("profile.confirmPass")} type="password" value={confirmPass}
                         onChange={(e) => setConfirmPass(e.target.value)} error={passError || undefined} />
              <Button onClick={submitPassword}
                      disabled={passwordMutation.isPending || !currentPass || !newPass}>
                {passwordMutation.isPending ? t("cf.saving") : t("profile.updatePass")}
              </Button>
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
