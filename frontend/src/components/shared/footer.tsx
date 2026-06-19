import Link from "next/link";
import { Logo } from "./logo";
import { ROUTES } from "@/constants";

const groups = [
  {
    title: "Platform",
    links: [
      { label: "Courses", href: ROUTES.courses },
      { label: "Teach", href: ROUTES.studio },
      { label: "My Learning", href: ROUTES.dashboard },
    ],
  },
  {
    title: "Company",
    links: [
      { label: "About", href: "#" },
      { label: "Careers", href: "#" },
      { label: "Blog", href: "#" },
    ],
  },
  {
    title: "Legal",
    links: [
      { label: "Privacy", href: "#" },
      { label: "Terms", href: "#" },
      { label: "Contact", href: "#" },
    ],
  },
];

export function Footer() {
  return (
    <footer className="border-t">
      <div className="mx-auto grid max-w-7xl gap-8 px-6 py-14 md:grid-cols-5">
        <div className="md:col-span-2">
          <Logo />
          <p className="mt-4 max-w-xs text-sm text-muted-foreground">
            The modern platform to learn, teach and grow. Built for the next generation of learners.
          </p>
        </div>
        {groups.map((group) => (
          <div key={group.title}>
            <h4 className="text-sm font-semibold">{group.title}</h4>
            <ul className="mt-3 space-y-2 text-sm text-muted-foreground">
              {group.links.map((link) => (
                <li key={link.label}>
                  <Link href={link.href} className="hover:text-foreground">
                    {link.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
      <div className="border-t py-6 text-center text-sm text-muted-foreground">
        © 2026 LearnHub. All rights reserved.
      </div>
    </footer>
  );
}
