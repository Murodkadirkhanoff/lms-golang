import { Navbar } from "@/components/shared/navbar";
import { Footer } from "@/components/shared/footer";
import { SkipLink } from "@/components/shared/skip-link";

export default function PublicLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen flex-col">
      <SkipLink />
      <Navbar />
      <main id="main-content" className="flex-1">
        {children}
      </main>
      <Footer />
    </div>
  );
}
