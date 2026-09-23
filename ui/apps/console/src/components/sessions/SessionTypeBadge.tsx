import type { ReactNode } from "react";
import { Badge, type BadgeShape } from "@shellhub/design-system/primitives";

import type { Session } from "@/client";
import { sessionType } from "@/utils/session";

interface SessionTypeBadgeProps {
  session: Session;
  shape?: BadgeShape;
  fallback?: ReactNode;
}

/**
 * The badge for what a session did. It renders the decision sessionType makes and makes none of its
 * own, so every screen that shows a session's kind agrees with every other.
 */
export default function SessionTypeBadge({
  session,
  shape = "pill",
  fallback = null,
}: SessionTypeBadgeProps) {
  const type = sessionType(session);
  if (!type) return fallback;

  return (
    <Badge color={type.color} shape={shape}>
      {type.label}
    </Badge>
  );
}
