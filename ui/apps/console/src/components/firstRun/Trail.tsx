import type { ReactNode } from "react";
import { CheckIcon } from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";

/**
 * Where a step stands: behind the user, in front of them now, or still ahead.
 */
export type TrailStepState = "done" | "active" | "upcoming";

/**
 * The vertical list of first-run steps, each one a TrailStep.
 */
export function Trail({ children }: { children: ReactNode }) {
  return <ol className="relative">{children}</ol>;
}

/**
 * The titles of the setup steps, the same on the setup page and on the dashboard it hands over
 * to.
 */
export const SETUP_STEP_TITLES = {
  survey: "Tell us about you",
  account: "Create your account and namespace",
} as const;

/**
 * The titles of the device steps, the same on every screen that lists them.
 */
export const DEVICE_STEP_TITLES = {
  install: "Install the agent",
  pair: "Check it's your device, then pair",
  shell: "Open a shell on it",
} as const;

/**
 * The device steps still ahead of a user who has not reached them, numbered from `start`.
 */
export function UpcomingDeviceSteps({ start }: { start: number }) {
  return (
    <>
      <TrailStep
        number={start}
        title={DEVICE_STEP_TITLES.install}
        state="upcoming"
      />
      <TrailStep
        number={start + 1}
        title={DEVICE_STEP_TITLES.pair}
        state="upcoming"
      />
      <TrailStep
        number={start + 2}
        title={DEVICE_STEP_TITLES.shell}
        state="upcoming"
      />
    </>
  );
}

interface TrailStepProps {
  number: number;
  title: ReactNode;
  state: TrailStepState;
  summary?: ReactNode;
  children?: ReactNode;
}

/**
 * One step of the trail. An active step shows its content, a done step collapses to its summary,
 * and an upcoming step shows only its title.
 */
export function TrailStep({
  number,
  title,
  state,
  summary,
  children,
}: TrailStepProps) {
  return (
    <li
      aria-current={state === "active" ? "step" : undefined}
      className="relative pl-11 pb-6 last:pb-0 group"
    >
      <span
        aria-hidden="true"
        className={cn(
          "absolute left-[13px] top-8 bottom-0 w-px bg-border group-last:hidden",
          state === "active" && "hidden sm:block",
        )}
      />
      <span
        aria-hidden="true"
        className={cn(
          "absolute left-0 top-0 w-7 h-7 rounded-full border flex items-center justify-center text-2xs font-mono font-semibold bg-background",
          state === "active" &&
            "border-primary text-primary shadow-[0_0_0_4px_rgba(102,122,204,0.12)]",
          state === "done" &&
            "border-accent-green bg-accent-green/10 text-accent-green",
          state === "upcoming" && "border-border-light text-text-muted",
        )}
      >
        {state === "done" ? (
          <CheckIcon className="w-3.5 h-3.5" strokeWidth={2.5} />
        ) : (
          number
        )}
      </span>
      <h2
        className={cn(
          "text-sm font-semibold leading-7",
          state === "active" ? "text-text-primary" : "text-text-secondary",
        )}
      >
        {title}
      </h2>
      {state === "done" && summary && (
        <p className="text-xs text-text-muted -mt-0.5">{summary}</p>
      )}
      {state === "active" && children && (
        <div className="mt-3 -ml-11 sm:ml-0 sm:mr-11">{children}</div>
      )}
    </li>
  );
}
