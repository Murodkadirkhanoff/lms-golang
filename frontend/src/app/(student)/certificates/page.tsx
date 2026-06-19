"use client";

import { useQuery } from "@tanstack/react-query";
import { Award, Download } from "lucide-react";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { LoadingState, EmptyState, ErrorState } from "@/components/shared/states";
import { dashboardService } from "@/services/dashboard.service";
import { formatDate } from "@/lib/utils";

export default function CertificatesPage() {
  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ["certificates"],
    queryFn: dashboardService.getCertificates,
  });

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-extrabold">Certificates</h1>
        <p className="text-muted-foreground">Your earned certificates of completion.</p>
      </div>

      {isLoading ? (
        <LoadingState />
      ) : isError ? (
        <ErrorState onRetry={() => refetch()} />
      ) : !data || data.length === 0 ? (
        <EmptyState title="No certificates yet" description="Complete a course to earn your first certificate." />
      ) : (
        <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {data.map((c) => (
            <Card key={c.id} className="overflow-hidden">
              <div className={`grid aspect-video place-items-center ${c.color}`}>
                <Award className="size-12 text-amber-700" />
              </div>
              <div className="p-5">
                <h3 className="font-bold">{c.courseTitle}</h3>
                <p className="mt-1 text-xs text-muted-foreground">Issued {formatDate(c.issuedAt)}</p>
                <Button variant="outline" size="sm" className="mt-4 w-full">
                  <Download className="size-4" /> Download PDF
                </Button>
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
