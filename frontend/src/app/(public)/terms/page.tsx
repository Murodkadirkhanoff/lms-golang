"use client";

import { LegalDoc } from "@/features/marketing/legal-doc";
import { useT } from "@/providers/locale-provider";

export default function TermsPage() {
  const t = useT();
  return (
    <LegalDoc
      title={t("terms.title")}
      updated={t("legal.updated")}
      intro={t("terms.intro")}
      sections={[
        { heading: t("terms.s1h"), body: [t("terms.s1b")] },
        { heading: t("terms.s2h"), body: [t("terms.s2b")] },
        { heading: t("terms.s3h"), body: [t("terms.s3b")] },
        { heading: t("terms.s4h"), body: [t("terms.s4b")] },
        { heading: t("terms.s5h"), body: [t("terms.s5b")] },
      ]}
    />
  );
}
