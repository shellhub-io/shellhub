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
import Alert from "@/components/common/Alert";
import Spinner from "@/components/common/Spinner";

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
      // Same vault, new home: lock so the next unlock reads from it.
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
    <div className="p-6">
      <div className="flex items-center gap-3 mb-4">
        <div className="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center">
          {toServer ? (
            <ServerStackIcon className="w-5 h-5 text-primary" />
          ) : (
            <ComputerDesktopIcon className="w-5 h-5 text-primary" />
          )}
        </div>
        <div>
          <h2
            id={titleId}
            className="text-base font-semibold text-text-primary"
          >
            {toServer
              ? "Sync vault to the ShellHub server"
              : "Move vault to this device"}
          </h2>
          <p className="text-2xs text-text-muted mt-0.5">
            {toServer
              ? "Use your keys from any machine"
              : "Keep your keys in this browser only"}
          </p>
        </div>
      </div>

      {checking ? (
        <div className="flex items-center justify-center gap-2 py-6 text-sm text-text-secondary">
          <Spinner />
          Checking the server
        </div>
      ) : toServer && conflict ? (
        <>
          <Alert variant="warning" className="mb-4">
            The ShellHub server already has a synced vault, created on another
            machine. Pick the vault to keep. The other one is deleted.
          </Alert>
          <div
            className="space-y-2"
            role="group"
            aria-label="Choose which vault to keep"
          >
            <button
              type="button"
              disabled={working}
              onClick={() => void run(() => adoptServerVault(scope))}
              className="w-full text-left px-4 py-3 rounded-lg border border-border hover:border-border-light hover:bg-hover-subtle transition-colors disabled:opacity-dim disabled:cursor-not-allowed"
            >
              <span className="block text-sm font-medium text-text-primary">
                Keep the synced vault
              </span>
              <span className="block text-2xs text-text-muted mt-0.5">
                Deletes the keys stored in this browser.
              </span>
            </button>
            <button
              type="button"
              disabled={working}
              onClick={() => void run(() => migrateLocalToServer(scope))}
              className="w-full text-left px-4 py-3 rounded-lg border border-border hover:border-border-light hover:bg-hover-subtle transition-colors disabled:opacity-dim disabled:cursor-not-allowed"
            >
              <span className="block text-sm font-medium text-text-primary">
                Keep this device's vault
              </span>
              <span className="block text-2xs text-text-muted mt-0.5">
                Replaces the synced vault on the server.
              </span>
            </button>
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
        <Alert variant="error" className="mt-4">
          {error}
        </Alert>
      )}

      <div className="flex justify-end gap-2 pt-4">
        <button
          type="button"
          onClick={onClose}
          disabled={working}
          className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
        >
          Cancel
        </button>
        {!checking && !(toServer && conflict) && (
          <button
            type="button"
            disabled={working || (toServer && !!error)}
            onClick={() =>
              void run(() =>
                toServer
                  ? migrateLocalToServer(scope)
                  : migrateServerToLocal(scope),
              )
            }
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
          >
            {working && <Spinner tone="onPrimary" />}
            {toServer ? "Sync vault" : "Move vault"}
          </button>
        )}
      </div>
    </div>
  );
}

export default function VaultSyncDialog({ open, onClose, direction }: Props) {
  const instanceId = useId();
  const titleId = `vault-sync-title-${instanceId}`;

  return (
    <BaseDialog
      open={open}
      onClose={onClose}
      size="sm"
      aria-labelledby={titleId}
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
