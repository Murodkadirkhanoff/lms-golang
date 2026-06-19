import * as React from "react";
import { cn } from "@/lib/utils";
import { LoadingState, EmptyState } from "./states";

export interface Column<T> {
  key: string;
  header: string;
  className?: string;
  align?: "left" | "right" | "center";
  render?: (row: T) => React.ReactNode;
  accessor?: (row: T) => React.ReactNode;
}

interface DataTableProps<T> {
  columns: Column<T>[];
  data: T[];
  isLoading?: boolean;
  emptyMessage?: string;
  rowKey: (row: T) => string | number;
  onRowClick?: (row: T) => void;
}

export function DataTable<T>({
  columns,
  data,
  isLoading,
  emptyMessage = "No records found",
  rowKey,
  onRowClick,
}: DataTableProps<T>) {
  const alignClass = (a?: "left" | "right" | "center") =>
    a === "right" ? "text-right" : a === "center" ? "text-center" : "text-left";

  return (
    <div className="overflow-hidden rounded-xl border bg-card">
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-secondary/60 text-muted-foreground">
            <tr>
              {columns.map((col) => (
                <th key={col.key} className={cn("px-6 py-3 font-medium", alignClass(col.align), col.className)}>
                  {col.header}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y">
            {!isLoading &&
              data.map((row) => (
                <tr
                  key={rowKey(row)}
                  onClick={() => onRowClick?.(row)}
                  className={cn("transition-colors hover:bg-secondary/50", onRowClick && "cursor-pointer")}
                >
                  {columns.map((col) => (
                    <td key={col.key} className={cn("px-6 py-4", alignClass(col.align))}>
                      {col.render ? col.render(row) : col.accessor ? col.accessor(row) : null}
                    </td>
                  ))}
                </tr>
              ))}
          </tbody>
        </table>
      </div>
      {isLoading && <LoadingState />}
      {!isLoading && data.length === 0 && <EmptyState description={emptyMessage} />}
    </div>
  );
}
