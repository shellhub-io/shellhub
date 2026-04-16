import type { ReactNode } from "react";
import { useHasPermission } from "@/hooks/useHasPermission";
import type { Action } from "@/utils/permission";

interface RestrictedActionProps {
  action: Action;
  children: ReactNode;
  /** Custom tooltip message. Defaults to the standard permission denial message. */
  message?: string;
}

/**
 * Wraps a UI action element (button, link, etc.) and prevents interaction when
 * the current user lacks the required permission for `action`.
 *
 * When restricted:
 * - Pointer events are blocked on the child (`pointer-events-none`)
 * - The outer wrapper gets `cursor-not-allowed`, `aria-disabled="true"`, and a
 *   native `title` tooltip explaining the restriction
 * - The child remains visible so users understand the feature exists
 *
 * When allowed: children are rendered without any wrapper.
 */
export default function RestrictedAction({
  action,
  children,
  message = "You don't have permission to perform this action.",
}: RestrictedActionProps) {
  const allowed = useHasPermission(action);

  if (allowed) {
    return <>{children}</>;
  }

  return (
    <span
      title={message}
      aria-disabled="true"
      className="inline-flex cursor-not-allowed"
    >
      {/* inert blocks all interaction (pointer and keyboard) with children */}
      <span className="pointer-events-none opacity-50" {...{ inert: "" }}>
        {children}
      </span>
    </span>
  );
}
