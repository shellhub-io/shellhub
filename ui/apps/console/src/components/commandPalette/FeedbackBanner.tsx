import {
  ExclamationTriangleIcon,
  CheckIcon,
} from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import type { Feedback } from "./items";

interface FeedbackBannerProps {
  feedback: Feedback | null;
}

/** Inline live region: assertive (`role="alert"`) for errors, polite
 *  (`role="status"`) for copy success. Render with `key={feedback?.text}` so a
 *  changed message remounts and re-announces. */
export default function FeedbackBanner({ feedback }: FeedbackBannerProps) {
  if (!feedback) return null;

  return (
    <div
      role={feedback.kind === "error" ? "alert" : "status"}
      className={cn(
        "flex items-center gap-2 px-4 py-2 text-xs border-b shrink-0",
        feedback.kind === "error"
          ? "text-accent-yellow bg-accent-yellow/10 border-accent-yellow/20"
          : "text-accent-green bg-accent-green/10 border-accent-green/20",
      )}
    >
      {feedback.kind === "error" ? (
        <ExclamationTriangleIcon
          className="w-4 h-4 shrink-0"
          aria-hidden="true"
        />
      ) : (
        <CheckIcon className="w-4 h-4 shrink-0" aria-hidden="true" />
      )}
      <span>{feedback.text}</span>
    </div>
  );
}
