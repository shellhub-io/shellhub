import type { ReactNode } from "react";
import { cn } from "@shellhub/design-system/cn";

/**
 * The card the sign-in, sign-up and recovery forms sit in.
 */
export default function LoginLayoutCard({
  className,
  children,
}: {
  className?: string;
  children: ReactNode;
}) {
  return (
    <div
      className={cn(
        "w-full max-w-md bg-card/80 border border-border rounded-2xl p-8 backdrop-blur-sm animate-slide-up",
        className,
      )}
    >
      {children}
    </div>
  );
}
