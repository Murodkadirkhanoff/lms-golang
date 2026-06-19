import type { Metadata } from "next";
import { Inter } from "next/font/google";
import { QueryProvider } from "@/providers/query-provider";
import { APP_NAME } from "@/constants";
import "./globals.css";

const inter = Inter({ subsets: ["latin"], variable: "--font-inter" });

export const metadata: Metadata = {
  title: {
    default: `${APP_NAME} — Learn without limits`,
    template: `%s · ${APP_NAME}`,
  },
  description:
    "Build in-demand skills with expert-led courses. Learn at your own pace, track progress, and earn certificates.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={`${inter.variable} font-sans`}>
        <QueryProvider>{children}</QueryProvider>
      </body>
    </html>
  );
}
