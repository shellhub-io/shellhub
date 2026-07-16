import { useState } from "react";
import {
  ArrowPathIcon,
  BoltIcon,
  TagIcon,
  TicketIcon,
} from "@heroicons/react/24/outline";
import { Button, Spinner } from "@shellhub/design-system/primitives";
import { useInstallKeys } from "@/hooks/useInstallKeys";
import { useUpdateInstallKey } from "@/hooks/useInstallKeyMutations";
import { usePaginatedListState } from "@/hooks/usePaginatedListState";
import { type InstallKey } from "@/client";
import PageHeader from "@/components/common/PageHeader";
import RestrictedAction from "@/components/common/RestrictedAction";
import EmptyState from "@/components/common/EmptyState";
import InstallKeysTable from "./InstallKeysTable";
import CreateInstallKeyDrawer from "./CreateInstallKeyDrawer";
import EditInstallKeyDrawer from "./EditInstallKeyDrawer";
import RevealInstallKeyDialog from "./RevealInstallKeyDialog";
import RevokeInstallKeyDialog from "./RevokeInstallKeyDialog";

const PER_PAGE = 10;

type InstallKeyListParams = {
  page: number;
};

const INSTALL_KEY_LIST_DEFAULTS: InstallKeyListParams = { page: 1 };

export default function InstallKeys() {
  const { params, setPage } = usePaginatedListState<InstallKeyListParams>({
    prefix: "installKey",
    defaults: INSTALL_KEY_LIST_DEFAULTS,
  });
  const page = params.page;
  const { installKeys, totalCount, isLoading } = useInstallKeys({ page });

  // The store pins the namespace's auto-managed legacy key first (its keyless-enrollment queue) and
  // paginates it together with the user's keys, so totalCount already counts it. Drive both the page
  // count and the item count off that one base, or the "N keys" label and the page controls disagree.
  const totalPages = Math.ceil(totalCount / PER_PAGE);
  const cards = installKeys;

  const updateKey = useUpdateInstallKey();
  const [createOpen, setCreateOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<InstallKey | null>(null);
  const [revealTarget, setRevealTarget] = useState<InstallKey | null>(null);
  const [revokeTarget, setRevokeTarget] = useState<InstallKey | null>(null);
  const [revokeConfirmText, setRevokeConfirmText] = useState("");
  const [revokeError, setRevokeError] = useState<string | null>(null);
  const [toggleError, setToggleError] = useState<string | null>(null);

  // Pause/resume is reversible, so it fires the mutation directly (no confirm).
  const toggleDisabled = async (key: InstallKey) => {
    setToggleError(null);
    try {
      await updateKey.mutateAsync({
        path: { key: key.name },
        body: { disabled: !key.disabled },
      });
    } catch (err) {
      setToggleError(
        err instanceof Error
          ? err.message
          : `Failed to ${key.disabled ? "enable" : "disable"} Install Key.`,
      );
    }
  };

  // Revoke is irreversible (no un-revoke) and cuts off the key for new enrollments, so it's gated
  // behind typing the key's name — mirroring the vault-reset confirmation.
  const openRevoke = (key: InstallKey) => {
    setRevokeConfirmText("");
    setRevokeError(null);
    setRevokeTarget(key);
  };

  const closeRevoke = () => {
    setRevokeConfirmText("");
    setRevokeError(null);
    setRevokeTarget(null);
  };

  const confirmRevoke = async () => {
    if (!revokeTarget) return;
    setRevokeError(null);
    try {
      await updateKey.mutateAsync({
        path: { key: revokeTarget.name },
        body: { revoked: true },
      });
      closeRevoke();
    } catch (err) {
      setRevokeError(
        err instanceof Error ? err.message : "Failed to revoke Install Key.",
      );
    }
  };

  // First run only when the namespace has no keys at all. The auto-managed legacy key always exists in
  // practice, so this hero is effectively a safety net: otherwise the table renders so the pinned legacy
  // (tenant-only) row — and the pending queue of keyless devices behind it — stays reachable even with
  // no user-created keys.
  if (!isLoading && totalCount === 0) {
    return (
      <>
        <EmptyState
          icon={<TicketIcon className="w-8 h-8" />}
          overline="Provisioning"
          title="Install Keys"
          description="Reusable, revocable credentials that register devices with your namespace. When an agent first connects with a key, the key's mode decides what happens: accept automatically, hold for review, ask your endpoint, or match a MAC allowlist."
          features={[
            {
              icon: <BoltIcon className="w-5 h-5" />,
              title: "Register on first connect",
              description:
                "A device joins your namespace the first time its agent reaches ShellHub, not when it is installed.",
            },
            {
              icon: <ArrowPathIcon className="w-5 h-5" />,
              title: "Reusable and revocable",
              description:
                "Register a single device or a whole fleet, then revoke the key to cut off new ones.",
            },
            {
              icon: <TagIcon className="w-5 h-5" />,
              title: "Tagged on arrival",
              description:
                "Apply namespace tags to every device that registers with the key.",
            },
          ]}
          footnote="Keys can expire, cap their usage, or make their devices ephemeral."
        >
          <RestrictedAction action="installKey:create">
            <Button
              size="lg"
              onClick={() => setCreateOpen(true)}
              icon={<TicketIcon className="w-4 h-4" strokeWidth={2} />}
            >
              Create your first key
            </Button>
          </RestrictedAction>
        </EmptyState>

        <CreateInstallKeyDrawer
          open={createOpen}
          onClose={() => setCreateOpen(false)}
        />
      </>
    );
  }

  return (
    <div>
      <PageHeader
        icon={<TicketIcon className="w-6 h-6" />}
        overline="Provisioning"
        title="Install Keys"
        description="Install keys are reusable credentials that register devices with your namespace. Each key's mode decides how a device is admitted."
      >
        <RestrictedAction action="installKey:create">
          <Button
            onClick={() => setCreateOpen(true)}
            icon={<TicketIcon className="w-4 h-4" strokeWidth={2} />}
          >
            Create Install Key
          </Button>
        </RestrictedAction>
      </PageHeader>

      {isLoading ? (
        <div className="flex justify-center py-24">
          <Spinner />
        </div>
      ) : (
        <div className="animate-fade-in">
          {toggleError && (
            <p className="mb-3 text-xs text-accent-red">{toggleError}</p>
          )}

          <InstallKeysTable
            data={cards}
            page={page}
            totalPages={totalPages}
            totalCount={totalCount}
            onPageChange={setPage}
            onReveal={setRevealTarget}
            onEdit={setEditTarget}
            onToggleDisabled={(k) => void toggleDisabled(k)}
            onRevoke={openRevoke}
          />
        </div>
      )}

      <CreateInstallKeyDrawer
        open={createOpen}
        onClose={() => setCreateOpen(false)}
      />
      <EditInstallKeyDrawer
        installKey={editTarget}
        onClose={() => setEditTarget(null)}
      />
      <RevealInstallKeyDialog
        installKey={revealTarget}
        onClose={() => setRevealTarget(null)}
      />
      <RevokeInstallKeyDialog
        installKey={revokeTarget}
        open={!!revokeTarget}
        confirmText={revokeConfirmText}
        onConfirmTextChange={setRevokeConfirmText}
        onClose={closeRevoke}
        onConfirm={confirmRevoke}
        error={revokeError}
      />
    </div>
  );
}
