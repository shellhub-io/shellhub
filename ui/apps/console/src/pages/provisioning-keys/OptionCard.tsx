import type { ReactNode } from "react";

/**
 * OptionCard frames one of a provisioning key's toggles, highlighted while the option is on.
 */
export default function OptionCard({
  active,
  children,
}: {
  active: boolean;
  children: ReactNode;
}) {
  return (
    <div
      className={`rounded-xl border p-4 transition-colors ${
        active
          ? "border-primary/40 bg-primary/[0.05]"
          : "border-border bg-transparent"
      }`}
    >
      {children}
    </div>
  );
}
