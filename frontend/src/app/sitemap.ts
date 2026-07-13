import type { MetadataRoute } from "next";
import { SITE_URL } from "@/constants";
import { coursesService } from "@/services/courses.service";

// Public, indexable routes only. Private/transactional areas are excluded
// (see robots.ts).
const STATIC_PATHS = [
  "",
  "/courses",
  "/categories",
  "/about",
  "/pricing",
  "/teach",
  "/help",
  "/contact",
  "/terms",
  "/privacy",
];

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const now = new Date();

  const staticEntries: MetadataRoute.Sitemap = STATIC_PATHS.map((path) => ({
    url: `${SITE_URL}${path}`,
    lastModified: now,
    changeFrequency: "weekly",
    priority: path === "" ? 1 : 0.7,
  }));

  let courseEntries: MetadataRoute.Sitemap = [];
  try {
    const { items } = await coursesService.list({ pageSize: 1000 });
    courseEntries = items.map((course) => ({
      url: `${SITE_URL}/courses/${course.slug}`,
      lastModified: new Date(course.createdAt),
      changeFrequency: "weekly",
      priority: 0.8,
    }));
  } catch {
    // If the catalog is unavailable, still return the static sitemap.
  }

  return [...staticEntries, ...courseEntries];
}
