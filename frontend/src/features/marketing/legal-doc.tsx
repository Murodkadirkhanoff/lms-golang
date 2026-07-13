export interface LegalSection {
  heading: string;
  body: string[];
}

export function LegalDoc({
  title,
  updated,
  intro,
  sections,
}: {
  title: string;
  updated: string;
  intro: string;
  sections: LegalSection[];
}) {
  return (
    <div className="mx-auto max-w-3xl px-6 py-14">
      <h1 className="text-3xl font-extrabold">{title}</h1>
      <p className="mt-2 text-sm text-muted-foreground">Last updated {updated}</p>
      <p className="mt-6 leading-relaxed text-muted-foreground">{intro}</p>
      <div className="mt-10 space-y-8">
        {sections.map((s, i) => (
          <section key={s.heading}>
            <h2 className="text-lg font-bold">
              {i + 1}. {s.heading}
            </h2>
            {s.body.map((p, j) => (
              <p key={j} className="mt-3 leading-relaxed text-muted-foreground">
                {p}
              </p>
            ))}
          </section>
        ))}
      </div>
    </div>
  );
}
