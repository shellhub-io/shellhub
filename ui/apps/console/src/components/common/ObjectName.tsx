import type { ReactNode } from "react";

/**
 * The name of the thing a dialog acts on, set apart in mono so a confirmation reads which key,
 * device or member it touches before the user agrees.
 */
export default function ObjectName({ children }: { children: ReactNode }) {
  return (
    <span className="inline-block max-w-full truncate align-bottom font-mono text-xs text-text-primary bg-background border border-border rounded px-1.5 py-px">
      {children}
    </span>
  );
}
