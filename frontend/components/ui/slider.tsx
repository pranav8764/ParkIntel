"use client";

import { cn } from "@/lib/utils";

interface SliderProps {
  value: number;
  min?: number;
  max?: number;
  step?: number;
  label?: string;
  displayValue?: string;
  onChange: (value: number) => void;
  className?: string;
  disabled?: boolean;
  id?: string;
}

function Slider({
  value,
  min = 0,
  max = 100,
  step = 1,
  label,
  displayValue,
  onChange,
  className,
  disabled,
  id,
}: SliderProps) {
  const percent = ((value - min) / (max - min)) * 100;

  return (
    <div className={cn("flex flex-col gap-2", className)}>
      {(label || displayValue) && (
        <div className="flex items-center justify-between">
          {label && (
            <label
              htmlFor={id}
              className="text-xs font-medium uppercase tracking-widest text-muted"
            >
              {label}
            </label>
          )}
          {displayValue && (
            <span className="mono-value text-sm text-gold font-semibold">
              {displayValue}
            </span>
          )}
        </div>
      )}
      <div className="relative h-2 w-full">
        <div className="absolute inset-0 rounded-full bg-white/5" />
        <div
          className="absolute inset-y-0 left-0 rounded-full bg-gradient-to-r from-gold/60 to-gold transition-all duration-150"
          style={{ width: `${percent}%` }}
        />
        <input
          id={id}
          type="range"
          min={min}
          max={max}
          step={step}
          value={value}
          onChange={(e) => onChange(Number(e.target.value))}
          disabled={disabled}
          className={cn(
            "absolute inset-0 w-full cursor-pointer appearance-none bg-transparent",
            "[&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:h-4 [&::-webkit-slider-thumb]:w-4",
            "[&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-gold [&::-webkit-slider-thumb]:border-2",
            "[&::-webkit-slider-thumb]:border-background [&::-webkit-slider-thumb]:shadow-lg [&::-webkit-slider-thumb]:shadow-gold/30",
            "[&::-webkit-slider-thumb]:transition-transform [&::-webkit-slider-thumb]:duration-150",
            "[&::-webkit-slider-thumb]:hover:scale-110",
            "[&::-moz-range-thumb]:h-4 [&::-moz-range-thumb]:w-4 [&::-moz-range-thumb]:rounded-full",
            "[&::-moz-range-thumb]:bg-gold [&::-moz-range-thumb]:border-2 [&::-moz-range-thumb]:border-background",
            "disabled:opacity-40 disabled:pointer-events-none"
          )}
        />
      </div>
    </div>
  );
}

export { Slider };
export type { SliderProps };
