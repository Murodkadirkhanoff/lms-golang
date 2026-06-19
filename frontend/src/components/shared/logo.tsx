import Link from "next/link";
import { cn } from "@/lib/utils";
import { APP_NAME, ROUTES } from "@/constants";

export function Logo({ className, light = false }: { className?: string; light?: boolean }) {
  return (
    <Link href={ROUTES.home} className={cn("flex items-center gap-2", className)}>
      <div className="grid size-9 place-items-center rounded-xl bg-primary font-bold text-primary-foreground">
        L
      </div>
      <span className={cn("text-lg font-extrabold tracking-tight", light && "text-white")}>{APP_NAME}</span>
    </Link>
  );
}
