import { useState } from "react";
import { Outlet } from "react-router-dom";
import {
  FingerPrintIcon,
  PlusIcon,
  KeyIcon,
  UserIcon,
  ShieldCheckIcon,
  BoltIcon,
  PencilSquareIcon,
  TrashIcon,
} from "@heroicons/react/24/outline";
import { Button, IconButton } from "@shellhub/design-system/primitives";
import { useSSHIdentities } from "@/hooks/useSSHIdentities";
import { useDeleteSSHIdentity } from "@/hooks/useSSHIdentityMutations";
import { useAuthStore } from "@/stores/authStore";
import type { SshIdentity } from "@/client";
import PageHeader from "@/components/common/PageHeader";
import EmptyState from "@/components/common/EmptyState";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import CopyButton from "@/components/common/CopyButton";
import DataTable, { type Column } from "@/components/common/DataTable";
import RestrictedAction from "@/components/common/RestrictedAction";
import { formatDate } from "@/utils/date";
import IdentityDrawer from "./IdentityDrawer";

export default function SSHIdentities() {
  const userId = useAuthStore((s) => s.userId);

  // Always request the namespace-wide list; the API returns every member's keys
  // to owners/admins and only the caller's own to everyone else, so the page
  // behaves the same either way (one User column, your own row marked).
  const { identities, isLoading } = useSSHIdentities(true);
  const deleteIdentity = useDeleteSSHIdentity();
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
      render: (i) => (
        <span className="text-sm font-medium text-text-primary">{i.name}</span>
      ),
    },
    {
      key: "user",
      header: "User",
      render: (i) => (
        <span className="inline-flex items-center gap-1.5 text-xs font-mono text-text-secondary">
          <UserIcon className="w-3 h-3 shrink-0" strokeWidth={2} />
          {i.user_name}
          {i.user_id === userId && (
            <span className="px-1.5 py-0.5 rounded text-2xs font-sans font-medium bg-primary/10 text-primary">
              you
            </span>
          )}
        </span>
      ),
    },
    {
      key: "fingerprint",
      header: "Fingerprint",
      render: (i) => (
        <div className="flex items-center gap-1">
          <code
            className="text-2xs font-mono text-text-muted truncate max-w-[240px]"
            title={i.fingerprint}
          >
            {i.fingerprint}
          </code>
          <CopyButton text={i.fingerprint} />
        </div>
      ),
    },
    {
      key: "added",
      header: "Added",
      render: (i) => (
        <span className="text-xs font-mono text-text-muted">
          {formatDate(i.created_at)}
        </span>
      ),
    },
    {
      key: "last-used",
      header: "Last used",
      render: (i) => (
        <span className="text-xs font-mono text-text-muted">
          {i.last_used_at ? formatDate(i.last_used_at) : "Never"}
        </span>
      ),
    },
    {
      key: "actions",
      header: "",
      headerClassName: "w-16",
      render: (i) => {
        const mine = i.user_id === userId;

        return (
          <div className="flex items-center justify-end gap-0.5">
            {/* Renaming is own-key only. */}
            {mine && (
              <RestrictedAction action="sshIdentity:enroll">
                <IconButton
                  variant="primary"
                  title="Rename"
                  aria-label={`Rename ${i.name}`}
                  onClick={() => openEdit(i)}
                >
                  <PencilSquareIcon className="w-4 h-4" strokeWidth={2} />
                </IconButton>
              </RestrictedAction>
            )}
            {/* Revoking your own key needs enroll; another member's needs manage. */}
            <RestrictedAction
              action={mine ? "sshIdentity:enroll" : "sshIdentity:manage"}
            >
              <IconButton
                variant="danger"
                title="Revoke"
                aria-label={`Revoke ${i.name}`}
                onClick={() => setDeleteTarget(i)}
              >
                <TrashIcon className="w-4 h-4" strokeWidth={2} />
              </IconButton>
            </RestrictedAction>
          </div>
        );
      },
    },
  ];

  /* Full-page onboarding empty state (no keys enrolled in the namespace). The
     Outlet still renders so an enroll modal can open over it — the first
     enrollment happens with an empty list. */
  if (!isLoading && identities.length === 0) {
    return (
      <>
        <EmptyState
          icon={<FingerPrintIcon className="w-8 h-8" />}
          overline="Identity Access"
          title="SSH Identities"
          description="Enrolled SSH keys are your identity under the identity access mode. Enrollment happens just-in-time on your first connection with a key, or you can pre-enroll one here."
          features={[
            {
              icon: <BoltIcon className="w-5 h-5" />,
              title: "Enroll Once",
              description:
                "Approve a key on first use; every later connection is recognized instantly.",
            },
            {
              icon: <KeyIcon className="w-5 h-5" />,
              title: "Key as Identity",
              description:
                "A stock OpenSSH client works — no wrapper, no per-session approval.",
            },
            {
              icon: <ShieldCheckIcon className="w-5 h-5" />,
              title: "Revocable",
              description:
                "Revoke a key to force re-enrollment on its next use.",
            },
          ]}
          footnote="Enrolled keys are scoped to this namespace."
        >
          <RestrictedAction action="sshIdentity:enroll">
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
        description="Manage the SSH keys enrolled as your identity under the identity access mode."
      >
        <RestrictedAction action="sshIdentity:enroll">
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
        emptyMessage="No keys enrolled in this namespace"
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
            ? The key will need to be enrolled again on its next use.
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
