import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Docker image uchun minimal server bundle (.next/standalone).
  output: "standalone",
  images: {
    remotePatterns: [{ protocol: "https", hostname: "**" }],
  },
};

export default nextConfig;
