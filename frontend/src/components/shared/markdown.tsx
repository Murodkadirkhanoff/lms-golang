import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { cn } from "@/lib/utils";

// Renders trusted markdown (lesson bodies). react-markdown does not render raw
// HTML by default, so untrusted input stays safe. Elements are mapped to Tailwind
// classes so we avoid a global typography plugin and keep theming consistent.
export function Markdown({ children, className }: { children: string; className?: string }) {
  return (
    <div className={cn("max-w-none text-sm leading-relaxed text-foreground", className)}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          h1: ({ children }) => <h1 className="mb-4 mt-6 text-2xl font-bold first:mt-0">{children}</h1>,
          h2: ({ children }) => <h2 className="mb-3 mt-6 text-xl font-bold first:mt-0">{children}</h2>,
          h3: ({ children }) => <h3 className="mb-2 mt-5 text-lg font-semibold first:mt-0">{children}</h3>,
          p: ({ children }) => <p className="mb-4 leading-7 text-muted-foreground">{children}</p>,
          ul: ({ children }) => <ul className="mb-4 ml-6 list-disc space-y-1 text-muted-foreground">{children}</ul>,
          ol: ({ children }) => <ol className="mb-4 ml-6 list-decimal space-y-1 text-muted-foreground">{children}</ol>,
          li: ({ children }) => <li className="leading-7">{children}</li>,
          a: ({ children, href }) => (
            <a href={href} target="_blank" rel="noopener noreferrer" className="font-medium text-primary underline underline-offset-2">
              {children}
            </a>
          ),
          strong: ({ children }) => <strong className="font-semibold text-foreground">{children}</strong>,
          blockquote: ({ children }) => (
            <blockquote className="mb-4 border-l-4 border-primary/40 pl-4 italic text-muted-foreground">{children}</blockquote>
          ),
          code: ({ className: c, children }) => {
            const isBlock = /language-/.test(c ?? "");
            return isBlock ? (
              <code className="block overflow-x-auto rounded-lg bg-slate-900 p-4 font-mono text-xs text-slate-100">{children}</code>
            ) : (
              <code className="rounded bg-secondary px-1.5 py-0.5 font-mono text-xs">{children}</code>
            );
          },
          pre: ({ children }) => <pre className="mb-4">{children}</pre>,
          hr: () => <hr className="my-6" />,
          table: ({ children }) => (
            <div className="mb-4 overflow-x-auto">
              <table className="w-full border-collapse text-left text-sm">{children}</table>
            </div>
          ),
          th: ({ children }) => <th className="border-b bg-secondary/50 px-3 py-2 font-semibold">{children}</th>,
          td: ({ children }) => <td className="border-b px-3 py-2 text-muted-foreground">{children}</td>,
        }}
      >
        {children}
      </ReactMarkdown>
    </div>
  );
}
