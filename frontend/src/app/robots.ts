import type { MetadataRoute } from "next";
import { SITE_URL } from "@/constants";

export default function robots(): MetadataRoute.Robots {
  return {
    rules: {
      userAgent: "*",
      allow: "/",
      // Keep private and transactional areas out of the index.
      disallow: [
        "/dashboard",
        "/my-courses",
        "/learn/",
        "/quiz/",
        "/certificates",
        "/purchases",
        "/notifications",
        "/profile",
        "/settings",
        "/wishlist",
        "/cart",
        "/checkout",
        "/studio",
        "/admin",
        "/login",
        "/register",
        "/forgot-password",
        "/reset-password",
      ],
    },
    sitemap: `${SITE_URL}/sitemap.xml`,
  };
}
