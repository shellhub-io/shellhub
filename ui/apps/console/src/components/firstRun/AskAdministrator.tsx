import { Button, Spinner } from "@shellhub/design-system/primitives";
import { useAuthStore } from "@/stores/authStore";
import { useNamespaceArrival } from "@/hooks/useNamespaceArrival";

/**
 * For a user whose account may not create namespaces: who to ask and what to give them. Once
 * they have been added, it takes them into the namespace by itself.
 */
export default function AskAdministrator() {
  const email = useAuthStore((s) => s.email);
  const { ready, enter, failed } = useNamespaceArrival({ autoEnter: true });

  return (
    <div className="space-y-4">
      <p className="text-xs text-text-secondary leading-relaxed">
        Your account can&apos;t create namespaces. Ask an administrator to add
        you to one, with this email:
      </p>
      <p className="w-fit font-mono text-sm text-text-primary bg-background border border-border rounded-md px-3 py-1.5">
        {email}
      </p>
      {failed ? (
        <div className="space-y-3">
          <p role="alert" className="text-xs text-accent-red">
            You were added, but getting into the namespace failed.
          </p>
          <Button onClick={enter}>Try again</Button>
        </div>
      ) : (
        <p className="flex items-center gap-2 text-xs text-text-muted">
          <Spinner size="sm" />
          {ready ? "You're in. Taking you there..." : "Waiting to be added..."}
        </p>
      )}
    </div>
  );
}
