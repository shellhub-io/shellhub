import { useState } from "react";
import { Outlet } from "react-router-dom";
import {
  FingerPrintIcon,
  PlusIcon,
  KeyIcon,
  CpuChipIcon,
  WindowIcon,
  ArrowUpTrayIcon,
  CommandLineIcon,
  ShieldCheckIcon,
  BoltIcon,
  PencilSquareIcon,
  TrashIcon,
  EllipsisVerticalIcon,
} from "@heroicons/react/24/outline";
import {
  Button,
  Dropdown,
  IconButton,
} from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { useListSshIdentities } from "@/client/api";
import { useDeleteSshIdentity } from "@/client/api";
import { useAuthStore } from "@/stores/authStore";
import type { SshIdentity } from "@/client";
import PageHeader from "@/components/common/PageHeader";
import EmptyState from "@/components/common/EmptyState";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import CopyButton from "@/components/common/CopyButton";
import DataTable, { type Column } from "@/components/common/DataTable";
import RestrictedAction from "@/components/common/RestrictedAction";
import { formatDateFull, formatDateShort } from "@/utils/date";
import UserBadge from "@/components/common/UserBadge";
import { useBrowserKeyFingerprint } from "@/hooks/useBrowserKey";
import {
  shortFingerprint,
  sshIdentityEndOfLife,
  sshIdentitySource,
  type IdentityStatusTone,
} from "@/utils/sshIdentity";
import IdentityDrawer from "./IdentityDrawer";

// Upload for a key somebody put here, a terminal for one an SSH client offered
// at login, a browser window for one a browser holds. The pair that carries the
// distinction is terminal against window, which is the difference between the
// two ways in.
const SOURCE_ICON = {
  manual: ArrowUpTrayIcon,
  approval: CommandLineIcon,
  browser: WindowIcon,
};

// Fixed tracks, not flex: the sub-labels have to land on the same x in every
// row, and a content-sized track puts them wherever that row's value ends.
const TIMELINE_GRID = "grid grid-cols-[6.5rem_6.5rem_6.5rem]";
const TIMELINE_LABEL =
  "text-2xs uppercase tracking-wider font-semibold text-text-muted/60";
const TIMELINE_VALUE = "text-xs text-text-muted tabular-nums whitespace-nowrap";
const EXPIRY_TONE: Record<IdentityStatusTone, string> = {
  dead: "text-accent-red",
  armed: "text-accent-yellow",
  quiet: "",
};

/**
 * The SSH identities page: the keys that may authenticate, whose they are, and how each will
 * end. Fetches every identity in the namespace, marking the caller's own.
 */
export default function SSHIdentities() {
  const userId = useAuthStore((s) => s.userId);

  const { data: identities = [], isLoading } = useListSshIdentities({
    all: true,
  });
  const browserKeyFingerprint = useBrowserKeyFingerprint();
  const isCurrentBrowser = (i: SshIdentity) =>
    i.source === "browser" && i.fingerprint === browserKeyFingerprint;
  const deleteIdentity = useDeleteSshIdentity();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<SshIdentity | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<SshIdentity | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  const closeDelete = () => {
    setDeleteError(null);
    setDeleteTarget(null);
  };

  const confirmDelete = async () => {
    if (!deleteTarget) return;
    setDeleteError(null);
    try {
      await deleteIdentity.mutateAsync({ path: { id: deleteTarget.id } });
      closeDelete();
    } catch (err) {
      setDeleteError(
        err instanceof Error ? err.message : "Failed to revoke key.",
      );
    }
  };

  const openNew = () => {
    setEditTarget(null);
    setDrawerOpen(true);
  };
  const openEdit = (identity: SshIdentity) => {
    setEditTarget(identity);
    setDrawerOpen(true);
  };
  const closeDrawer = () => {
    setDrawerOpen(false);
    setEditTarget(null);
  };

  const columns: Column<SshIdentity>[] = [
    {
      key: "name",
      header: "Name",
      render: (i) => {
        const isService = i.principal_type === "service";
        const source = sshIdentitySource(i.source, isCurrentBrowser(i));
        const SourceIcon = SOURCE_ICON[i.source ?? "manual"] ?? KeyIcon;

        return (
          <div className="flex items-center gap-3">
            <span
              className={cn(
                "grid place-items-center w-8 h-8 rounded-lg shrink-0",
                isService
                  ? "bg-accent-yellow/10 text-accent-yellow"
                  : "bg-primary/10 text-primary",
              )}
            >
              {isService ? (
                <CpuChipIcon className="w-4 h-4" strokeWidth={2} />
              ) : (
                <KeyIcon className="w-4 h-4" strokeWidth={2} />
              )}
            </span>

            <div className="flex flex-col gap-0.5 min-w-0">
              <span className="text-sm font-medium text-text-primary">
                {i.name}
              </span>

              <span className="flex items-center gap-1.5 text-2xs text-text-muted">
                {/* Colour carries the only distinction worth a glance: the key
                    this browser holds is primary, every other origin is muted.
                    The label is the icon's accessible name, so it is still
                    readable without seeing the colour or the shape. */}
                <span
                  role="img"
                  aria-label={source.label}
                  title={source.description}
                  className={cn(
                    "shrink-0",
                    source.notable ? "text-primary" : "text-text-muted",
                  )}
                >
                  <SourceIcon className="w-3.5 h-3.5" strokeWidth={2} />
                </span>
                <code className="font-mono" title={i.fingerprint}>
                  {shortFingerprint(i.fingerprint)}
                </code>
                <CopyButton text={i.fingerprint} />
              </span>
            </div>
          </div>
        );
      },
    },
    {
      key: "user",
      header: "Principal",
      render: (i) => {
        const isService = i.principal_type === "service";

        return (
          <UserBadge
            name={i.principal_name}
            email={isService ? undefined : i.principal_email}
            secondary={isService ? "Service account" : undefined}
            trailing={
              !isService && i.principal_id === userId ? (
                <span className="px-1.5 py-0.5 rounded text-2xs font-medium bg-primary/10 text-primary">
                  you
                </span>
              ) : undefined
            }
          />
        );
      },
    },
    {
      key: "timeline",
      header: "Timeline",
      render: (i) => {
        const end = sshIdentityEndOfLife(i);

        return (
          <dl className={TIMELINE_GRID}>
            <div className="flex flex-col gap-0.5">
              <dt className={TIMELINE_LABEL}>Added</dt>
              <dd
                className={TIMELINE_VALUE}
                title={formatDateFull(i.created_at)}
              >
                {formatDateShort(i.created_at)}
              </dd>
            </div>
            <div className="flex flex-col gap-0.5">
              <dt className={TIMELINE_LABEL}>Used</dt>
              <dd
                className={cn(TIMELINE_VALUE, !i.last_used_at && "italic")}
                title={
                  i.last_used_at ? formatDateFull(i.last_used_at) : undefined
                }
              >
                {i.last_used_at ? formatDateShort(i.last_used_at) : "never"}
              </dd>
            </div>
            <div className="flex flex-col gap-0.5">
              <dt className={TIMELINE_LABEL}>Expires</dt>
              <dd
                className={cn(
                  TIMELINE_VALUE,
                  EXPIRY_TONE[end.tone],
                  end.kind === "never" && "italic",
                )}
                title={end.title}
              >
                {end.value}
              </dd>
            </div>
          </dl>
        );
      },
    },
    {
      key: "actions",
      header: "",
      headerClassName: "w-16",
      render: (i) => {
        const mine = i.principal_id === userId;

        return (
          <div className="flex justify-end">
            <Dropdown portal placement="bottom-end">
              <Dropdown.Trigger>
                <IconButton
                  variant="ghost"
                  aria-label={`Actions for ${i.name}`}
                >
                  <EllipsisVerticalIcon className="w-4 h-4" />
                </IconButton>
              </Dropdown.Trigger>

              <Dropdown.Panel className="w-40 py-1">
                {/* Editing is own-key only, and renaming is all it does. */}
                {mine && (
                  <RestrictedAction action="sshIdentity:add">
                    <Dropdown.Item
                      label="Edit"
                      onSelect={() => openEdit(i)}
                      className="gap-2.5 px-3 py-2"
                    >
                      <PencilSquareIcon className="w-4 h-4" />
                      Edit
                    </Dropdown.Item>
                  </RestrictedAction>
                )}
                {/* Revoking your own key needs add; another member's needs manage. */}
                <RestrictedAction
                  action={mine ? "sshIdentity:add" : "sshIdentity:manage"}
                >
                  <Dropdown.Item
                    label="Revoke"
                    variant="danger"
                    onSelect={() => setDeleteTarget(i)}
                    className="gap-2.5 px-3 py-2"
                  >
                    <TrashIcon className="w-4 h-4" />
                    Revoke
                  </Dropdown.Item>
                </RestrictedAction>
              </Dropdown.Panel>
            </Dropdown>
          </div>
        );
      },
    },
  ];

  if (!isLoading && identities.length === 0) {
    return (
      <>
        <EmptyState
          icon={<FingerPrintIcon className="w-8 h-8" />}
          overline="Identity Access"
          title="SSH Identities"
          description="Under the identity access mode an SSH key is who you are. A key becomes an identity the first time you connect with it, or you can add one here ahead of time."
          features={[
            {
              icon: <BoltIcon className="w-5 h-5" />,
              title: "Approve Once",
              description:
                "Approve a key on first use; every later connection is recognized instantly.",
            },
            {
              icon: <KeyIcon className="w-5 h-5" />,
              title: "Key as Identity",
              description:
                "A stock OpenSSH client works — no wrapper, no approval on every login.",
            },
            {
              icon: <ShieldCheckIcon className="w-5 h-5" />,
              title: "Revocable",
              description:
                "Revoke a key and its next use has to be approved again.",
            },
          ]}
          footnote="Identities are scoped to this namespace."
        >
          <RestrictedAction action="sshIdentity:add">
            <Button
              size="lg"
              onClick={openNew}
              icon={<PlusIcon className="w-4 h-4" strokeWidth={2} />}
            >
              Add a key
            </Button>
          </RestrictedAction>
        </EmptyState>

        <IdentityDrawer
          open={drawerOpen}
          editIdentity={editTarget}
          onClose={closeDrawer}
        />

        <Outlet />
      </>
    );
  }

  return (
    <div>
      <PageHeader
        icon={<FingerPrintIcon className="w-6 h-6" />}
        overline="Security"
        title="SSH Identities"
        description="Manage the SSH keys that are your identity under the identity access mode."
      >
        <RestrictedAction action="sshIdentity:add">
          <Button
            onClick={openNew}
            icon={<PlusIcon className="w-4 h-4" strokeWidth={2} />}
          >
            Add a key
          </Button>
        </RestrictedAction>
      </PageHeader>

      <DataTable
        columns={columns}
        data={identities}
        rowKey={(i) => i.id}
        isLoading={isLoading}
        loadingMessage="Loading SSH identities..."
        emptyMessage="No SSH identities in this namespace"
      />

      <IdentityDrawer
        open={drawerOpen}
        editIdentity={editTarget}
        onClose={closeDrawer}
      />

      <ConfirmDialog
        open={!!deleteTarget}
        onClose={closeDelete}
        onConfirm={confirmDelete}
        title="Revoke Key"
        description={
          <>
            Are you sure you want to revoke{" "}
            <span className="font-medium text-text-primary">
              {deleteTarget?.name}
            </span>
            ?{" "}
            {
              sshIdentitySource(
                deleteTarget?.source,
                !!deleteTarget && isCurrentBrowser(deleteTarget),
              ).revokeConsequence
            }
          </>
        }
        confirmLabel="Revoke"
      >
        {deleteError && (
          <p className="text-xs text-accent-red">{deleteError}</p>
        )}
      </ConfirmDialog>

      <Outlet />
    </div>
  );
}
