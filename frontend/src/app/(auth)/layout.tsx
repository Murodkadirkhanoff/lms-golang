import Link from "next/link";
import { Logo } from "@/components/shared/logo";
import { ROUTES } from "@/constants";

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="grid min-h-screen lg:grid-cols-2">
      {/* Brand side */}
      <div className="relative hidden flex-col justify-between bg-gradient-to-br from-primary to-violet-700 p-12 text-primary-foreground lg:flex">
        <Logo light />
        <div>
          <h2 className="text-4xl font-extrabold leading-tight">Learn without limits.</h2>
          <p className="mt-4 max-w-sm text-indigo-100">
            Join 320,000+ learners. Take courses, track progress, earn certificates — and teach your own courses too.
          </p>
        </div>
        <p className="text-sm text-indigo-200">© 2026 LearnHub</p>
      </div>

      {/* Form side */}
      <div className="flex flex-col items-center justify-center p-6">
        <div className="w-full max-w-sm">
          <div className="mb-8 lg:hidden">
            <Logo />
          </div>
          {children}
          <p className="mt-8 text-center text-xs text-muted-foreground">
            By continuing you agree to our{" "}
            <Link href="#" className="underline">
              Terms
            </Link>{" "}
            and{" "}
            <Link href="#" className="underline">
              Privacy Policy
            </Link>
            .
          </p>
        </div>
      </div>
    </div>
  );
}
