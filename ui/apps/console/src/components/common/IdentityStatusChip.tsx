import { cn } from "@shellhub/design-system/cn";
import {
  sshIdentityStatus,
  type IdentityStatusTone,
} from "@/utils/sshIdentity";
import type { SshIdentity } from "@/client";

const TONE: Record<IdentityStatusTone, string> = {
  dead: "bg-text-muted/10 text-text-muted",
  armed: "bg-accent-yellow/10 text-accent-yellow",
  // Deliberately not primary: that colour marks the key this browser holds, and
  // a far-off expiry must not compete with it.
  quiet: "bg-text-muted/10 text-text-muted",
};

/** The lifecycle chip for an enrolled key: when it expires, or that it is
 *  already spent. Renders nothing for a durable key (no TTL, reusable, never
 *  consumed), which is the default. */
export default function IdentityStatusChip({
  identity,
}: {
  identity: Pick<SshIdentity, "expires_at" | "consumed_at" | "single_use">;
}) {
  const status = sshIdentityStatus(identity);
  if (!status) return null;

  return (
    <span
      title={status.title}
      className={cn(
        "px-1.5 py-0.5 rounded text-2xs font-medium whitespace-nowrap",
        TONE[status.tone],
      )}
    >
      {status.label}
    </span>
  );
}
