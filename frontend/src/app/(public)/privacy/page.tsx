"use client";

import { LegalDoc } from "@/features/marketing/legal-doc";
import { useT } from "@/providers/locale-provider";

export default function PrivacyPage() {
  const t = useT();
  return (
    <LegalDoc
      title={t("privacy.title")}
      updated={t("legal.updated")}
      intro={t("privacy.intro")}
      sections={[
        { heading: t("privacy.s1h"), body: [t("privacy.s1b")] },
        { heading: t("privacy.s2h"), body: [t("privacy.s2b")] },
        { heading: t("privacy.s3h"), body: [t("privacy.s3b")] },
        { heading: t("privacy.s4h"), body: [t("privacy.s4b")] },
        { heading: t("privacy.s5h"), body: [t("privacy.s5b")] },
      ]}
    />
  );
}
