import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

/**
 * Merge Tailwind class names safely.
 * Combines clsx for conditional logic with tailwind-merge for dedup.
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
