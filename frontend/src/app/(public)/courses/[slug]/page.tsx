import type { Metadata } from "next";
import { CourseDetail } from "@/features/courses/course-detail";
import { coursesService } from "@/services/courses.service";
import { APP_NAME, SITE_URL } from "@/constants";

type Params = { params: Promise<{ slug: string }> };

export async function generateMetadata({ params }: Params): Promise<Metadata> {
  const { slug } = await params;
  try {
    const course = await coursesService.getBySlug(slug);
    const description = course.description.slice(0, 160);
    const url = `${SITE_URL}/courses/${course.slug}`;
    return {
      title: course.title,
      description,
      alternates: { canonical: `/courses/${course.slug}` },
      openGraph: {
        type: "website",
        title: `${course.title} · ${APP_NAME}`,
        description,
        url,
      },
      twitter: { card: "summary_large_image", title: course.title, description },
    };
  } catch {
    return { title: "Course not found" };
  }
}

export default async function CoursePage({ params }: Params) {
  const { slug } = await params;

  // Structured data helps search engines surface the course as a rich result.
  let jsonLd: string | null = null;
  try {
    const course = await coursesService.getBySlug(slug);
    jsonLd = JSON.stringify({
      "@context": "https://schema.org",
      "@type": "Course",
      name: course.title,
      description: course.description,
      provider: { "@type": "Organization", name: APP_NAME, sameAs: SITE_URL },
      instructor: { "@type": "Person", name: course.instructor.name },
      ...(course.ratingCount > 0 && {
        aggregateRating: {
          "@type": "AggregateRating",
          ratingValue: course.rating,
          ratingCount: course.ratingCount,
        },
      }),
      offers: {
        "@type": "Offer",
        price: course.price,
        priceCurrency: "UZS",
        availability: "https://schema.org/InStock",
        url: `${SITE_URL}/courses/${course.slug}`,
      },
    });
  } catch {
    // No structured data when the course can't be loaded.
  }

  return (
    <>
      {jsonLd && (
        <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: jsonLd }} />
      )}
      <CourseDetail slug={slug} />
    </>
  );
}
