import { KeyIcon, UserIcon } from "@heroicons/react/24/outline";
import { Badge } from "@shellhub/design-system/primitives";

import type { SessionPrincipal as Principal } from "@/client";

interface SessionPrincipalProps {
  principal: Principal | undefined;
  /**
   * The account's name, resolved by the caller from the members and API keys it already loaded.
   * Without it the chip shows a truncated id, with the whole id on the element's title.
   */
  name?: string;
}

const ICON = "w-3 h-3 shrink-0";

/**
 * Who opened the session, a person or an API key. It renders nothing when the session has no
 * principal, which is every session under the legacy access model.
 */
export default function SessionPrincipal({
  principal,
  name,
}: SessionPrincipalProps) {
  if (!principal) return null;

  const isAutomation = principal.kind === "api-key";

  return (
    <Badge color={isAutomation ? "yellow" : "primary"} shape="pill">
      {isAutomation ? (
        <KeyIcon className={ICON} strokeWidth={2} />
      ) : (
        <UserIcon className={ICON} strokeWidth={2} />
      )}
      {name ?? (
        <span className="font-mono" title={principal.id}>
          {`${principal.id.slice(0, 12)}…`}
        </span>
      )}
    </Badge>
  );
}
