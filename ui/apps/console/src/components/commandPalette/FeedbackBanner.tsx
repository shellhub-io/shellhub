import { ExclamationTriangleIcon } from "@heroicons/react/24/outline";

interface FeedbackBannerProps {
  message: string | null;
}

/** Inline live region (`role="alert"`) for rejected actions — e.g. connecting
 *  to an offline device or without permission. Render with `key={message}` so a
 *  changed message remounts and re-announces. Null when there's nothing to say. */
export default function FeedbackBanner({ message }: FeedbackBannerProps) {
  if (!message) return null;

  return (
    <div
      role="alert"
      className="flex items-center gap-2 px-4 py-2 text-xs text-accent-yellow bg-accent-yellow/10 border-b border-accent-yellow/20 shrink-0"
    >
      <ExclamationTriangleIcon
        className="w-4 h-4 shrink-0"
        aria-hidden="true"
      />
      <span>{message}</span>
    </div>
  );
}
