"use client";

import { Card } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";

const engagement = [
  { course: "Complete Next.js 16 Course", completion: 68, students: 2340 },
  { course: "React Patterns & Performance", completion: 54, students: 1820 },
  { course: "Advanced TypeScript Deep Dive", completion: 72, students: 1440 },
];

export default function StudioAnalyticsPage() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-extrabold">Student analytics</h1>

      <div className="grid gap-4 sm:grid-cols-3">
        <Card className="p-5">
          <span className="text-sm text-muted-foreground">Active students</span>
          <div className="mt-2 text-3xl font-extrabold">8,420</div>
        </Card>
        <Card className="p-5">
          <span className="text-sm text-muted-foreground">Avg. completion</span>
          <div className="mt-2 text-3xl font-extrabold">64%</div>
        </Card>
        <Card className="p-5">
          <span className="text-sm text-muted-foreground">Avg. quiz score</span>
          <div className="mt-2 text-3xl font-extrabold">81%</div>
        </Card>
      </div>

      <Card className="p-6">
        <h2 className="mb-4 text-lg font-bold">Completion by course</h2>
        <div className="space-y-5">
          {engagement.map((e) => (
            <div key={e.course}>
              <div className="mb-1 flex justify-between text-sm">
                <span className="font-medium">{e.course}</span>
                <span className="text-muted-foreground">{e.completion}% · {e.students} students</span>
              </div>
              <Progress value={e.completion} />
            </div>
          ))}
        </div>
      </Card>
    </div>
  );
}
