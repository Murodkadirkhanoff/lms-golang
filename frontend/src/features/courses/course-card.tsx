import Link from "next/link";
import { Star } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { ROUTES } from "@/constants";
import { formatNumber, formatPrice } from "@/lib/utils";
import type { Course } from "@/types";

export function CourseCard({ course }: { course: Course }) {
  return (
    <Link href={ROUTES.course(course.slug)} className="group">
      <Card className="overflow-hidden transition-shadow hover:shadow-lg">
        <div className={`aspect-video bg-gradient-to-br ${course.thumbnailColor}`} />
        <div className="space-y-2 p-5">
          <Badge>{course.category}</Badge>
          <h3 className="line-clamp-2 font-bold leading-snug group-hover:text-primary">{course.title}</h3>
          <p className="text-sm text-muted-foreground">by {course.instructor.name}</p>
          <div className="flex items-center gap-1 text-sm">
            <Star className="size-4 fill-amber-400 text-amber-400" />
            <span className="font-semibold">{course.rating}</span>
            <span className="text-muted-foreground">({formatNumber(course.ratingCount)})</span>
          </div>
          <div className="pt-1 text-lg font-extrabold">{formatPrice(course.price)}</div>
        </div>
      </Card>
    </Link>
  );
}
