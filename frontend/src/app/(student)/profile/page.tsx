"use client";

import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

const learningStats = [
  { label: "Hours learned", value: "142h" },
  { label: "Courses completed", value: "5" },
  { label: "Current streak", value: "7 days" },
  { label: "Quizzes passed", value: "23" },
];

export default function ProfilePage() {
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
          <TabsTrigger value="info">Personal info</TabsTrigger>
          <TabsTrigger value="password">Password</TabsTrigger>
          <TabsTrigger value="stats">Statistics</TabsTrigger>
        </TabsList>

        <TabsContent value="info">
          <Card>
            <CardHeader>
              <CardTitle>Personal information</CardTitle>
              <CardDescription>Update your account details.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 sm:grid-cols-2">
                <FormField label="Full name" defaultValue="Amir Karimov" />
                <FormField label="Email" type="email" defaultValue="amir@mail.com" />
              </div>
              <FormField label="Headline" defaultValue="Lifelong learner" />
              <Button>Save changes</Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="password">
          <Card>
            <CardHeader>
              <CardTitle>Change password</CardTitle>
              <CardDescription>Use a strong, unique password.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <FormField label="Current password" type="password" />
              <FormField label="New password" type="password" />
              <FormField label="Confirm new password" type="password" />
              <Button>Update password</Button>
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
