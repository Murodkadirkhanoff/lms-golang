"use client";

import { useRef, useState } from "react";
import { Bold, Italic, Heading, List, ListOrdered, Link2, Code } from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Markdown } from "@/components/shared/markdown";
import { cn } from "@/lib/utils";
import { useT } from "@/providers/locale-provider";

interface MarkdownEditorProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
}

// A dependency-light markdown editor: a textarea with a formatting toolbar and a
// live Preview tab (rendered by the shared <Markdown>). Selection-aware so the
// toolbar wraps the current selection. Controlled via value/onChange so it plugs
// into react-hook-form.
export function MarkdownEditor({ value, onChange, placeholder }: MarkdownEditorProps) {
  const t = useT();
  const ref = useRef<HTMLTextAreaElement>(null);

  // Wraps the current selection with `before`/`after`, or inserts a line prefix.
  function apply(before: string, after = before, blockPrefix?: string) {
    const el = ref.current;
    if (!el) return;
    const start = el.selectionStart;
    const end = el.selectionEnd;
    const selected = value.slice(start, end);

    let next: string;
    let cursor: number;
    if (blockPrefix) {
      next = `${value.slice(0, start)}${blockPrefix}${selected || ""}${value.slice(end)}`;
      cursor = start + blockPrefix.length + selected.length;
    } else {
      next = `${value.slice(0, start)}${before}${selected || ""}${after}${value.slice(end)}`;
      cursor = selected ? start + before.length + selected.length + after.length : start + before.length;
    }
    onChange(next);
    requestAnimationFrame(() => {
      el.focus();
      el.setSelectionRange(cursor, cursor);
    });
  }

  const tools = [
    { icon: Bold, label: "Bold", run: () => apply("**") },
    { icon: Italic, label: "Italic", run: () => apply("_") },
    { icon: Heading, label: "Heading", run: () => apply("", "", "## ") },
    { icon: List, label: "Bullet list", run: () => apply("", "", "- ") },
    { icon: ListOrdered, label: "Numbered list", run: () => apply("", "", "1. ") },
    { icon: Link2, label: "Link", run: () => apply("[", "](https://)") },
    { icon: Code, label: "Code", run: () => apply("`") },
  ];

  const [tab, setTab] = useState("write");

  return (
    <div className="overflow-hidden rounded-lg border">
      <Tabs value={tab} onValueChange={setTab}>
        <div className="flex items-center justify-between gap-2 border-b bg-secondary/50 px-2 py-1.5">
          <div className="flex items-center gap-0.5">
            {tools.map((tool) => (
              <button
                key={tool.label}
                type="button"
                title={tool.label}
                aria-label={tool.label}
                disabled={tab !== "write"}
                onClick={tool.run}
                className="grid size-7 place-items-center rounded text-muted-foreground hover:bg-background hover:text-foreground disabled:opacity-40"
              >
                <tool.icon className="size-4" />
              </button>
            ))}
          </div>
          <TabsList className="h-7">
            <TabsTrigger value="write" className="text-xs">{t("md.write")}</TabsTrigger>
            <TabsTrigger value="preview" className="text-xs">{t("md.preview")}</TabsTrigger>
          </TabsList>
        </div>

        <TabsContent value="write" className="m-0">
          <textarea
            ref={ref}
            value={value}
            onChange={(e) => onChange(e.target.value)}
            placeholder={placeholder ?? t("md.placeholder")}
            rows={12}
            className={cn(
              "block w-full resize-y bg-background px-4 py-3 font-mono text-sm leading-6",
              "focus:outline-none",
            )}
          />
        </TabsContent>

        <TabsContent value="preview" className="m-0">
          <div className="min-h-[18rem] px-4 py-3">
            {value.trim() ? <Markdown>{value}</Markdown> : <p className="text-sm text-muted-foreground">{t("md.empty")}</p>}
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
