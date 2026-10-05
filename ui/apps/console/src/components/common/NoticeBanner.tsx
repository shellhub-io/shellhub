import { ReactNode } from "react";
import { cn } from "@shellhub/design-system/cn";

/**
 * Props of NoticeBanner. visible is a prop rather than a mount condition so the banner can
 * animate out instead of vanishing.
 */
export interface NoticeBannerProps {
  visible: boolean;
  severity: "error" | "warning";
  align?: "start" | "center";
  children: ReactNode;
}

type SeverityConfig = {
  surface: string;
  text: string;
  dot: string;
  role: "alert" | "status";
  ariaLive: "assertive" | "polite";
};

const SEVERITY: Record<"error" | "warning", SeverityConfig> = {
  error: {
    surface: "bg-accent-red/[0.06] border-accent-red/10",
    text: "text-accent-red",
    dot: "bg-accent-red",
    role: "alert",
    ariaLive: "assertive",
  },
  warning: {
    surface: "bg-accent-yellow/[0.06] border-accent-yellow/10",
    text: "text-accent-yellow",
    dot: "bg-accent-yellow",
    role: "status",
    ariaLive: "polite",
  },
};

/**
 * The page-level banner for something the user should know but need not act on now.
 */
export default function NoticeBanner({
  visible,
  severity,
  align = "start",
  children,
}: NoticeBannerProps) {
  const { surface, text, dot, role, ariaLive } = SEVERITY[severity];
  const justifyClass = align === "center" ? "justify-center" : "justify-start";

  return (
    <div
      aria-hidden={!visible ? true : undefined}
      {...(!visible ? { inert: true } : {})}
      className={cn(
        "grid transition-[grid-template-rows] duration-300 ease-out",
        visible ? "grid-rows-[1fr]" : "grid-rows-[0fr]",
      )}
    >
      <div className="overflow-hidden">
        <div
          role={role}
          aria-live={ariaLive}
          className={cn(surface, text, justifyClass, "px-5 py-1.5 flex items-center gap-2 border-b")}
        >
          {visible && (
            <>
              <span
                className={cn("inline-flex rounded-full h-1.5 w-1.5 shrink-0", dot)}
              />
              <p className="text-xs font-mono">{children}</p>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
