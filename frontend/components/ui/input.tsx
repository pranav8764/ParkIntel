"use client";

import { forwardRef, type InputHTMLAttributes } from "react";
import { cn } from "@/lib/utils";

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
}

const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, label, id, ...props }, ref) => {
    return (
      <div className="flex flex-col gap-1.5">
        {label && (
          <label
            htmlFor={id}
            className="text-xs font-medium uppercase tracking-widest text-muted"
          >
            {label}
          </label>
        )}
        <input
          ref={ref}
          id={id}
          className={cn(
            "h-10 w-full rounded-lg px-3",
            "bg-white/5 text-foreground placeholder:text-muted/60",
            "border border-white/10",
            "transition-all duration-300",
            "hover:border-white/15",
            "focus:outline-none focus:border-gold/40 focus:ring-2 focus:ring-gold/15",
            "disabled:opacity-40 disabled:pointer-events-none",
            "text-sm font-mono",
            className
          )}
          {...props}
        />
      </div>
    );
  }
);

Input.displayName = "Input";

export { Input };
export type { InputProps };
