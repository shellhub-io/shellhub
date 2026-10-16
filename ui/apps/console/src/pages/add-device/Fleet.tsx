import { useState } from "react";
import { PlusIcon } from "@heroicons/react/24/outline";
import {
  Button,
  Callout,
  Spinner,
  WindowChrome,
} from "@shellhub/design-system/primitives";
import CopyButton from "@/components/common/CopyButton";
import { type ProvisioningKey } from "@/client";
import RadioGroupField from "@/components/common/fields/RadioGroupField";
import RadioPill from "@/components/common/fields/RadioPill";
import RestrictedAction from "@/components/common/RestrictedAction";
import { useHasPermission } from "@/hooks/useHasPermission";
import { usePaginatedListState } from "@/hooks/usePaginatedListState";
import { useProvisioningKeys } from "@/hooks/useProvisioningKeys";
import { useRevealProvisioningKey } from "@/hooks/useRevealProvisioningKey";
import { methodsForNewNamespace, type Method } from "@/pages/install/methods";
import CreateProvisioningKeyModal from "@/pages/provisioning-keys/CreateProvisioningKeyModal";
import EditProvisioningKeyModal from "@/pages/provisioning-keys/EditProvisioningKeyModal";
import ProvisioningKeysTable from "@/pages/provisioning-keys/ProvisioningKeysTable";
import RevokeProvisioningKeyDialog from "@/pages/provisioning-keys/RevokeProvisioningKeyDialog";
import {
  isInstallable,
  isLegacyKey,
  isSystemKey,
  keyModeInfo,
} from "@/pages/provisioning-keys/helpers";
import { useToggleProvisioningKey } from "@/pages/provisioning-keys/useToggleProvisioningKey";
import { buildConnectorCommand } from "@/utils/installCommand";
import { pageCount } from "@/utils/pagination";
import InstallCommand from "./InstallCommand";

type ProvisioningKeyListParams = {
  page: number;
};

type FleetMethod = Method | "connector";

function ConnectorCommand({ provisioningKey }: { provisioningKey: string }) {
  const command = buildConnectorCommand(
    `PROVISIONING_KEY=${provisioningKey}`,
    window.location.origin,
  );

  return (
    <div className="space-y-3">
      <WindowChrome
        variant="terminal"
        size="sm"
        title="Run on the Docker host"
        titleBarSlot={<CopyButton text={command} showLabel />}
      >
        <pre className="text-xs leading-relaxed whitespace-pre-wrap [overflow-wrap:anywhere]">
          <span className="text-text-muted select-none">$ </span>
          <span className="text-accent-cyan">{command}</span>
        </pre>
      </WindowChrome>
      <p className="text-xs text-text-muted leading-relaxed">
        The connector enrolls every container on the host with this key, and
        each one is decided the way the key says.
      </p>
    </div>
  );
}

const PROVISIONING_KEY_LIST_DEFAULTS: ProvisioningKeyListParams = { page: 1 };

/**
 * Provisioning many devices unattended: the namespace's provisioning keys, all folded until one is
 * clicked, which opens its install command, or the Docker connector command, under its row and
 * folds again on a second click. The legacy key is listed only while it has pending devices. When
 * the key's secret cannot be loaded, or the viewer's role may not read it, the row says so instead
 * of showing a command without it; the secret is only requested for a role that may read it.
 */
export default function Fleet() {
  const { params, setPage } = usePaginatedListState<ProvisioningKeyListParams>({
    prefix: "provisioningKey",
    defaults: PROVISIONING_KEY_LIST_DEFAULTS,
  });
  const page = params.page;
  const { provisioningKeys, totalCount, isLoading } = useProvisioningKeys({
    page,
  });
  const totalPages = pageCount(totalCount);

  const [openName, setOpenName] = useState<string | null>(null);
  const [method, setMethod] = useState<FleetMethod>("auto");
  const [createOpen, setCreateOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<ProvisioningKey | null>(null);
  const [revokeTarget, setRevokeTarget] = useState<ProvisioningKey | null>(
    null,
  );
  const { toggle, error: toggleError } = useToggleProvisioningKey();

  const listed = provisioningKeys.filter(
    (key) => !isLegacyKey(key) || (key.pending_devices ?? 0) > 0,
  );
  const listedCount = totalCount - (provisioningKeys.length - listed.length);
  const installable = listed.filter(isInstallable);
  const selectedKey = installable.find((k) => k.name === openName);
  const canRevealKey = useHasPermission("provisioningKey:reveal");
  const { key: revealedKey, error: revealError } = useRevealProvisioningKey(
    selectedKey?.name ?? null,
    canRevealKey,
  );

  const noCustomKeys =
    !provisioningKeys.some((key) => !isSystemKey(key)) && totalPages <= 1;

  return (
    <div className="space-y-6">
      <section className="space-y-3">
        <div className="flex items-start justify-between gap-4">
          <div>
            <h2 className="text-sm font-semibold text-text-primary">
              Provisioning keys
            </h2>
            <p className="mt-0.5 text-xs text-text-muted">
              Pick the key to bake into your image. Its mode decides whether
              each device gets in.
            </p>
          </div>
          <RestrictedAction action="provisioningKey:create">
            <Button
              size="sm"
              onClick={() => setCreateOpen(true)}
              icon={<PlusIcon className="w-4 h-4" strokeWidth={2} />}
            >
              New key
            </Button>
          </RestrictedAction>
        </div>

        {isLoading ? (
          <div className="flex justify-center py-24">
            <Spinner />
          </div>
        ) : (
          <div className="animate-fade-in">
            {toggleError && (
              <p className="mb-3 text-xs text-accent-red">{toggleError}</p>
            )}
            <ProvisioningKeysTable
              data={listed}
              selectedName={selectedKey?.name}
              onSelect={(k) =>
                setOpenName(k.name === selectedKey?.name ? null : k.name)
              }
              renderInstall={(k) => (
                <div className="border-t border-primary/15 bg-primary/[0.03] px-5 py-4 space-y-4">
                  <RadioGroupField
                    label="Installation method"
                    value={method}
                    onChange={setMethod}
                    containerClassName="flex flex-wrap gap-1.5 mt-2"
                  >
                    {methodsForNewNamespace().map((m) => (
                      <RadioPill key={m.id} value={m.id} label={m.label} />
                    ))}
                    <RadioPill value="connector" label="Docker connector" />
                  </RadioGroupField>
                  {!canRevealKey ? (
                    <Callout variant="info">
                      Your role cannot read this key, so there is no command to
                      copy. Ask an administrator for it.
                    </Callout>
                  ) : revealError ? (
                    <Callout variant="error">
                      Could not load this key, so there is no command to copy.
                      Check your connection and try again.
                    </Callout>
                  ) : method === "connector" ? (
                    <ConnectorCommand provisioningKey={revealedKey || "…"} />
                  ) : (
                    <InstallCommand
                      method={method}
                      credential={`PROVISIONING_KEY=${revealedKey || "…"}`}
                      outcome={
                        <>
                          Each machine that runs this command enrolls with{" "}
                          <span className="font-medium text-text-primary">
                            {k.name}
                          </span>{" "}
                          and is {keyModeInfo(k).outcome}.
                        </>
                      }
                    />
                  )}
                </div>
              )}
              page={page}
              totalPages={totalPages}
              totalCount={listedCount}
              noCustomKeys={noCustomKeys}
              onPageChange={setPage}
              onCreate={() => setCreateOpen(true)}
              onEdit={setEditTarget}
              onToggleDisabled={(k) => void toggle(k)}
              onRevoke={setRevokeTarget}
            />
          </div>
        )}
      </section>

      <CreateProvisioningKeyModal
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onCreated={(name) => setOpenName(name)}
      />
      <EditProvisioningKeyModal
        provisioningKey={editTarget}
        onClose={() => setEditTarget(null)}
      />
      <RevokeProvisioningKeyDialog
        provisioningKey={revokeTarget}
        onRevoked={() => setRevokeTarget(null)}
      />
    </div>
  );
}
