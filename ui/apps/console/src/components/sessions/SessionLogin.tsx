import { ExclamationTriangleIcon } from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";

import type { Session } from "@/client";

/**
 * The account on the device a session logged in as, flagged when the session never
 * authenticated.
 */
export default function SessionLogin({ session }: { session: Session }) {
  const suspicious = !session.authenticated;

  return (
    <div className="flex items-center gap-1.5">
      {suspicious && (
        <ExclamationTriangleIcon
          className="w-3.5 h-3.5 text-accent-red/70 shrink-0"
          strokeWidth={2}
          title="Not authenticated"
        />
      )}
      <code
        className={cn(
          "text-xs font-mono bg-surface px-1.5 py-0.5 rounded",
          suspicious ? "text-accent-red/60" : "text-text-muted",
        )}
      >
        {session.username}
      </code>
    </div>
  );
}
