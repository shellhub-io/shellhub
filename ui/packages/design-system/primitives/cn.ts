import { clsx, type ClassValue } from "clsx";
import { extendTailwindMerge } from "tailwind-merge";

const twMerge = extendTailwindMerge({
  extend: {
    classGroups: {
      "font-size": [{ text: ["2xs", "3xs"] }],
    },
  },
});

/**
 * Joins class names and resolves Tailwind conflicts, the last one winning, so a caller can
 * override any class a component sets without knowing the order they were written in.
 * The merge is taught the design-system's own text-2xs and text-3xs sizes, which Tailwind
 * does not ship and would otherwise not recognise as part of the font-size group.
 */
export function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs));
}
