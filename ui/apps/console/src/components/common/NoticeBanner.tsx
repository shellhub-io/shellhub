import { ReactNode } from "react";

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
      {...(!visible ? { inert: "" } : {})}
      className={`grid transition-[grid-template-rows] duration-300 ease-out ${
        visible ? "grid-rows-[1fr]" : "grid-rows-[0fr]"
      }`}
    >
      <div className="overflow-hidden">
        {/*
         * The strip container is always mounted so the overflow-hidden clip
         * layer has content to collapse into, producing the smooth
         * grid-rows 1fr→0fr transition on hide. The live region (role +
         * aria-live) is registered on this always-present element so the
         * browser tracks it from the start.
         *
         * Children are conditionally rendered: screen readers fire live-region
         * announcements on content insertion, NOT on aria-hidden removal. By
         * mounting children only when visible=true, the text node is freshly
         * inserted into the registered live region each time the banner
         * becomes visible, which reliably triggers an AT announcement.
         *
         * aria-hidden + inert on the outer wrapper suppress AT access when
         * the banner is not visible; focusable children are also absent from
         * the DOM at that point.
         */}
        <div
          role={role}
          aria-live={ariaLive}
          className={`${surface} ${text} ${justifyClass} px-5 py-1.5 flex items-center gap-2 border-b`}
        >
          {visible && (
            <>
              <span
                className={`inline-flex rounded-full h-1.5 w-1.5 shrink-0 ${dot}`}
              />
              <p className="text-xs font-mono">{children}</p>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
