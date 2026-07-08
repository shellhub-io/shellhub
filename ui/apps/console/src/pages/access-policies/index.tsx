import { useState } from "react";
import {
  ShieldCheckIcon,
  PlusIcon,
  TagIcon,
  UsersIcon,
  UserIcon,
  GlobeAltIcon,
  ServerIcon,
  CommandLineIcon,
  PencilSquareIcon,
  TrashIcon,
  IdentificationIcon,
  ExclamationTriangleIcon,
} from "@heroicons/react/24/outline";
import { Button, IconButton } from "@shellhub/design-system/primitives";
import { useAccessPolicies } from "@/hooks/useAccessPolicies";
import { useDeleteAccessPolicy } from "@/hooks/useAccessPolicyMutations";
import { useNamespace } from "@/hooks/useNamespaces";
import { useAuthStore } from "@/stores/authStore";
import type { AccessPolicy } from "@/client";
import PageHeader from "@/components/common/PageHeader";
import EmptyState from "@/components/common/EmptyState";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import DataTable, { type Column } from "@/components/common/DataTable";
import RestrictedAction from "@/components/common/RestrictedAction";
import { formatDate } from "@/utils/date";
import AccessPolicyDrawer from "./AccessPolicyDrawer";

/* ── subject cell ─────────────────────────────────── */

function SubjectCell({ policy }: { policy: AccessPolicy }) {
  const { type, value } = policy.subject;
  if (type === "all-members") {
    return (
      <span className="inline-flex items-center gap-1.5 text-xs font-mono text-text-secondary">
        <UsersIcon className="w-3 h-3 shrink-0" strokeWidth={2} />
        All members
      </span>
    );
  }
  if (type === "role") {
    return (
      <span className="inline-flex items-center gap-1.5 text-xs font-mono text-text-secondary">
        <IdentificationIcon className="w-3 h-3 shrink-0" strokeWidth={2} />
        Role: {value}
      </span>
    );
  }
  return (
    <span className="inline-flex items-center gap-1.5 text-xs font-mono text-text-secondary">
      <UserIcon className="w-3 h-3 shrink-0" strokeWidth={2} />
      {value}
    </span>
  );
}

/* ── devices cell ─────────────────────────────────── */

function DevicesCell({ policy }: { policy: AccessPolicy }) {
  if (policy.filter.tags.length > 0) {
    return (
      <span className="inline-flex items-center gap-1.5 flex-wrap">
        {policy.filter.tags.map((tag) => (
          <span
            key={tag.id}
            className="inline-flex items-center gap-1 px-1.5 py-0.5 bg-primary/10 text-primary text-2xs font-mono rounded"
          >
            <TagIcon className="w-2.5 h-2.5 shrink-0" strokeWidth={2} />
            {tag.name}
          </span>
        ))}
      </span>
    );
  }
  if (policy.filter.hostname && policy.filter.hostname !== ".*") {
    return (
      <span className="inline-flex items-center gap-1 px-1.5 py-0.5 bg-accent-yellow/10 text-accent-yellow text-2xs font-mono rounded">
        <ServerIcon className="w-2.5 h-2.5 shrink-0" strokeWidth={2} />
        {policy.filter.hostname}
      </span>
    );
  }
  return (
    <span className="inline-flex items-center gap-1 text-2xs font-mono text-text-muted">
      <GlobeAltIcon className="w-2.5 h-2.5 shrink-0" strokeWidth={2} />
      All devices
    </span>
  );
}

/* ── logins cell ──────────────────────────────────── */

function LoginsCell({ policy }: { policy: AccessPolicy }) {
  const isAny = policy.logins.length === 1 && policy.logins[0] === "*";
  if (isAny) {
    return (
      <span className="inline-flex items-center gap-1 text-2xs font-mono text-text-muted">
        <CommandLineIcon className="w-2.5 h-2.5 shrink-0" strokeWidth={2} />
        Any login
      </span>
    );
  }
  return (
    <span className="inline-flex items-center gap-1.5 flex-wrap">
      {policy.logins.map((login) => (
        <span
          key={login}
          className="inline-flex items-center gap-1 px-1.5 py-0.5 bg-accent-cyan/10 text-accent-cyan text-2xs font-mono rounded"
        >
          {login}
        </span>
      ))}
    </span>
  );
}

/* ── page ─────────────────────────────────────────── */

export default function AccessPolicies() {
  const { policies, isLoading } = useAccessPolicies();
  const { tenant: tenantId } = useAuthStore();
  const { namespace: ns } = useNamespace(tenantId ?? "");
  const isIdentityMode = ns?.settings?.ssh_access_mode === "identity";
  const deletePolicy = useDeleteAccessPolicy();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<AccessPolicy | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<AccessPolicy | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  const closeDelete = () => {
    setDeleteError(null);
    setDeleteTarget(null);
  };

  const confirmDelete = async () => {
    if (!deleteTarget) return;
    setDeleteError(null);
    try {
      await deletePolicy.mutateAsync({ path: { id: deleteTarget.id } });
      closeDelete();
    } catch (err) {
      setDeleteError(
        err instanceof Error ? err.message : "Failed to delete access policy.",
      );
    }
  };

  const openNew = () => {
    setEditTarget(null);
    setDrawerOpen(true);
  };
  const openEdit = (policy: AccessPolicy) => {
    setEditTarget(policy);
    setDrawerOpen(true);
  };
  const closeDrawer = () => {
    setDrawerOpen(false);
    setEditTarget(null);
  };

  const columns: Column<AccessPolicy>[] = [
    {
      key: "name",
      header: "Name",
      render: (p) => (
        <span className="text-sm font-medium text-text-primary">{p.name}</span>
      ),
    },
    {
      key: "subject",
      header: "Subject",
      render: (p) => <SubjectCell policy={p} />,
    },
    {
      key: "devices",
      header: "Devices",
      render: (p) => <DevicesCell policy={p} />,
    },
    {
      key: "logins",
      header: "Logins",
      render: (p) => <LoginsCell policy={p} />,
    },
    {
      key: "created",
      header: "Created",
      render: (p) => (
        <span className="text-xs font-mono text-text-muted">
          {formatDate(p.created_at)}
        </span>
      ),
    },
    {
      key: "actions",
      header: "",
      headerClassName: "w-16",
      render: (p) => (
        <div className="flex items-center justify-end gap-0.5">
          <RestrictedAction action="accessPolicy:edit">
            <IconButton
              variant="primary"
              title="Edit"
              aria-label={`Edit ${p.name}`}
              onClick={() => openEdit(p)}
            >
              <PencilSquareIcon className="w-4 h-4" strokeWidth={2} />
            </IconButton>
          </RestrictedAction>
          <RestrictedAction action="accessPolicy:remove">
            <IconButton
              variant="danger"
              title="Delete"
              aria-label={`Delete ${p.name}`}
              onClick={() => setDeleteTarget(p)}
            >
              <TrashIcon className="w-4 h-4" strokeWidth={2} />
            </IconButton>
          </RestrictedAction>
        </div>
      ),
    },
  ];

  /* Full-page onboarding empty state (no policies at all) */
  if (!isLoading && policies.length === 0) {
    return (
      <>
        <EmptyState
          accent={isIdentityMode ? "yellow" : "primary"}
          icon={
            isIdentityMode ? (
              <ExclamationTriangleIcon className="w-8 h-8" />
            ) : (
              <ShieldCheckIcon className="w-8 h-8" />
            )
          }
          overline={isIdentityMode ? "Access Blocked" : "Identity Access"}
          title={
            isIdentityMode ? "SSH is blocked for everyone" : "Access Policies"
          }
          description={
            isIdentityMode
              ? "Identity mode is on but this namespace has no policies, so every SSH login is denied. Add a policy to allow access."
              : "Decide who may reach which devices, as which login, under the identity SSH access mode. Policies are allow-only and default-deny."
          }
          features={[
            {
              icon: <UsersIcon className="w-5 h-5" />,
              title: "Subject-based",
              description:
                "Grant access to all members, a role, or a single member.",
            },
            {
              icon: <TagIcon className="w-5 h-5" />,
              title: "Device Scoping",
              description: "Target devices by hostname pattern or by tags.",
            },
            {
              icon: <CommandLineIcon className="w-5 h-5" />,
              title: "Login Control",
              description:
                "Allow any login, or restrict to an explicit list of unix logins.",
            },
          ]}
          footnote={
            isIdentityMode
              ? undefined
              : "Policies take effect once you switch SSH access to identity mode in Settings."
          }
        >
          <RestrictedAction action="accessPolicy:create">
            <Button
              size="lg"
              onClick={openNew}
              icon={<PlusIcon className="w-4 h-4" strokeWidth={2} />}
            >
              {isIdentityMode ? "Add policy" : "Add your first policy"}
            </Button>
          </RestrictedAction>
        </EmptyState>

        <AccessPolicyDrawer
          open={drawerOpen}
          editPolicy={editTarget}
          onClose={closeDrawer}
        />
      </>
    );
  }

  return (
    <div>
      <PageHeader
        icon={<ShieldCheckIcon className="w-6 h-6" />}
        overline="Security"
        title="Access Policies"
        description="Control who may reach which devices, as which login, under the identity SSH access mode."
      >
        <RestrictedAction action="accessPolicy:create">
          <Button
            onClick={openNew}
            icon={<PlusIcon className="w-4 h-4" strokeWidth={2} />}
          >
            Add Policy
          </Button>
        </RestrictedAction>
      </PageHeader>

      <DataTable
        columns={columns}
        data={policies}
        rowKey={(p) => p.id}
        isLoading={isLoading}
        loadingMessage="Loading access policies..."
        emptyMessage="No access policies found"
      />

      <AccessPolicyDrawer
        open={drawerOpen}
        editPolicy={editTarget}
        onClose={closeDrawer}
      />

      <ConfirmDialog
        open={!!deleteTarget}
        onClose={closeDelete}
        onConfirm={confirmDelete}
        title="Delete Access Policy"
        description={
          <>
            Are you sure you want to delete{" "}
            <span className="font-medium text-text-primary">
              {deleteTarget?.name}
            </span>
            ? This action cannot be undone.
          </>
        }
        confirmLabel="Delete"
      >
        {deleteError && (
          <p className="text-xs text-accent-red">{deleteError}</p>
        )}
      </ConfirmDialog>
    </div>
  );
}
