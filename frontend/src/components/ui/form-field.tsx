import * as React from "react";
import { Label } from "./label";
import { Input } from "./input";
import { cn } from "@/lib/utils";

interface FormFieldProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label: string;
  error?: string;
}

export const FormField = React.forwardRef<HTMLInputElement, FormFieldProps>(
  ({ label, error, id, className, ...props }, ref) => {
    const fieldId = id ?? props.name;
    const errorId = error ? `${fieldId}-error` : undefined;
    return (
      <div className="space-y-1.5">
        <Label htmlFor={fieldId}>{label}</Label>
        <Input
          id={fieldId}
          ref={ref}
          className={cn(error && "border-destructive focus-visible:ring-destructive", className)}
          aria-invalid={!!error}
          aria-describedby={errorId}
          {...props}
        />
        {error && (
          <p id={errorId} role="alert" className="text-xs text-destructive">
            {error}
          </p>
        )}
      </div>
    );
  },
);
FormField.displayName = "FormField";
