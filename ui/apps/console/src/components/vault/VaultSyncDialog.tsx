import { useState, useEffect, useId, useMemo } from "react";
import {
  ServerStackIcon,
  ComputerDesktopIcon,
} from "@heroicons/react/24/outline";
import { useVaultStore } from "@/stores/vaultStore";
import { useAuthStore } from "@/stores/authStore";
import {
  serverVaultExists,
  migrateLocalToServer,
  adoptServerVault,
  migrateServerToLocal,
} from "@/utils/vault-migrate";
import BaseDialog from "@/components/common/BaseDialog";
import { Button, Callout, Spinner } from "@shellhub/design-system/primitives";
import DialogHeader from "@/components/common/DialogHeader";

type Direction = "to-server" | "to-local";

interface Props {
  open: boolean;
  onClose: () => void;
  direction: Direction;
}

function useScope() {
  const user = useAuthStore((s) => s.user);
  const tenant = useAuthStore((s) => s.tenant);
  return useMemo(
    () => (user && tenant ? { user, tenant } : undefined),
    [user, tenant],
  );
}

function SyncForm({
  open,
  onClose,
  direction,
  instanceId,
}: Props & { instanceId: string }) {
  const refreshStatus = useVaultStore((s) => s.refreshStatus);
  const lock = useVaultStore((s) => s.lock);
  const scope = useScope();

  const [checking, setChecking] = useState(direction === "to-server");
  const [conflict, setConflict] = useState(false);
  const [working, setWorking] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!open || direction !== "to-server") return undefined;
    let cancelled = false;
    serverVaultExists(scope)
      .then((exists) => {
        if (!cancelled) setConflict(exists);
      })
      .catch(() => {
        if (!cancelled) setError("Could not reach the server. Try again.");
      })
      .finally(() => {
        if (!cancelled) setChecking(false);
      });
    return () => {
      cancelled = true;
    };
  }, [open, direction, scope]);

  const run = async (action: () => Promise<void>) => {
    setWorking(true);
    setError(null);
    try {
      await action();
      lock();
      await refreshStatus();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to move the vault");
    } finally {
      setWorking(false);
    }
  };

  const titleId = `vault-sync-title-${instanceId}`;
  const toServer = direction === "to-server";

  return (
    <div>
      <DialogHeader
        icon={toServer ? <ServerStackIcon /> : <ComputerDesktopIcon />}
        title={
          toServer
            ? "Sync vault to the ShellHub server"
            : "Move vault to this device"
        }
        description={
          toServer
            ? "Use your keys from any machine you sign in on."
            : "Keep your keys in this browser only."
        }
        titleId={titleId}
        descriptionId={`${titleId}-description`}
        onClose={working ? undefined : onClose}
      />
      <div className="px-6 pb-6">

      {checking ? (
        <div className="flex items-center justify-center gap-2 py-6 text-sm text-text-secondary">
          <Spinner />
          Checking the server
        </div>
      ) : toServer && conflict ? (
        <>
          <Callout variant="warning" className="mb-4">
            The ShellHub server already has a synced vault, created on another
            machine. Pick the vault to keep. The other one is deleted.
          </Callout>
          <div
            className="space-y-2"
            role="group"
            aria-label="Choose which vault to keep"
          >
            <Button
              variant="outline"
              fullWidth
              className="flex-col items-start justify-start"
              disabled={working}
              onClick={() => void run(() => adoptServerVault(scope))}
            >
              <span className="block text-sm font-medium text-text-primary">
                Keep the synced vault
              </span>
              <span className="block text-2xs text-text-muted mt-0.5">
                Deletes the keys stored in this browser.
              </span>
            </Button>
            <Button
              variant="outline"
              fullWidth
              className="flex-col items-start justify-start"
              disabled={working}
              onClick={() => void run(() => migrateLocalToServer(scope))}
            >
              <span className="block text-sm font-medium text-text-primary">
                Keep this device's vault
              </span>
              <span className="block text-2xs text-text-muted mt-0.5">
                Replaces the synced vault on the server.
              </span>
            </Button>
          </div>
        </>
      ) : (
        <div className="text-sm text-text-secondary space-y-3 mb-1">
          {toServer ? (
            <>
              <p>
                Your encrypted vault moves to the ShellHub server and is removed
                from this browser. Unlock it with the same master password from
                any machine you sign in to.
              </p>
              <p className="text-xs text-text-muted">
                Encryption stays in your browser. The server never sees your
                keys or your master password.
              </p>
            </>
          ) : (
            <>
              <p>
                Your encrypted vault moves to this browser and is removed from
                the ShellHub server. Other machines lose access to it.
              </p>
              <p className="text-xs text-text-muted">
                Clearing this browser's data deletes the vault permanently.
              </p>
            </>
          )}
        </div>
      )}

      {error && (
        <Callout variant="error" className="mt-4">
          {error}
        </Callout>
      )}

      <div className="flex justify-end gap-2 pt-4">
        <Button variant="ghost" onClick={onClose} disabled={working}>
          Cancel
        </Button>
        {!checking && !(toServer && conflict) && (
          <Button
            disabled={toServer && !!error}
            loading={working}
            onClick={() =>
              void run(() =>
                toServer
                  ? migrateLocalToServer(scope)
                  : migrateServerToLocal(scope),
              )
            }
          >
            {toServer ? "Sync vault" : "Move vault"}
          </Button>
        )}
      </div>
      </div>
    </div>
  );
}

/**
 * Moves a vault between this browser and the server. direction names which way; the copy at the
 * destination is replaced, so the dialog has to say which side wins before it runs.
 */
export default function VaultSyncDialog({ open, onClose, direction }: Props) {
  const instanceId = useId();
  const titleId = `vault-sync-title-${instanceId}`;

  return (
    <BaseDialog
      open={open}
      onClose={onClose}
      size="sm"
      aria-labelledby={titleId}
      aria-describedby={`${titleId}-description`}
    >
      <SyncForm
        key={String(open)}
        open={open}
        onClose={onClose}
        direction={direction}
        instanceId={instanceId}
      />
    </BaseDialog>
  );
}
