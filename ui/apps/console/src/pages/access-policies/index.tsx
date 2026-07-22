import { useState, type ReactNode } from "react";
import {
  ShieldCheckIcon,
  PlusIcon,
  MagnifyingGlassIcon,
  TagIcon,
  UsersIcon,
  UserIcon,
  CpuChipIcon,
  GlobeAltIcon,
  ServerIcon,
  CommandLineIcon,
  PencilSquareIcon,
  TrashIcon,
  IdentificationIcon,
  ExclamationTriangleIcon,
  CheckCircleIcon,
  NoSymbolIcon,
} from "@heroicons/react/24/outline";
import { Button, IconButton } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { useAccessPolicies } from "@/hooks/useAccessPolicies";
import { useDeleteAccessPolicy } from "@/hooks/useAccessPolicyMutations";
import { useNamespace } from "@/hooks/useNamespaces";
import { useServiceAccounts } from "@/hooks/useServiceAccounts";
import { useAuthStore } from "@/stores/authStore";
import type { AccessPolicy } from "@/client";
import PageHeader from "@/components/common/PageHeader";
import EmptyState from "@/components/common/EmptyState";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import DataTable, { type Column } from "@/components/common/DataTable";
import RestrictedAction from "@/components/common/RestrictedAction";
import { formatDate } from "@/utils/date";
import AccessPolicyDrawer from "./AccessPolicyDrawer";

/* ── cell chip ────────────────────────────────────── */

const CHIP_TONE = {
  neutral: "bg-card text-text-secondary border border-border",
  primary: "bg-primary/10 text-primary",
  cyan: "bg-accent-cyan/10 text-accent-cyan",
  green: "bg-accent-green/10 text-accent-green",
  red: "bg-accent-red/10 text-accent-red",
} as const;

const CHIP_ICON = "w-3 h-3 shrink-0";

function Chip({
  icon,
  tone = "neutral",
  mono,
  title,
  children,
}: {
  icon?: ReactNode;
  tone?: keyof typeof CHIP_TONE;
  mono?: boolean;
  title?: string;
  children: ReactNode;
}) {
  return (
    <span
      title={title}
      className={cn(
        "inline-flex items-center gap-1.5 px-2 py-1 rounded-md text-xs font-medium",
        CHIP_TONE[tone],
        mono && "font-mono",
      )}
    >
      {icon}
      {children}
    </span>
  );
}

/* ── subject cell ─────────────────────────────────── */

function SubjectCell({
  policy,
  memberEmail,
  serviceAccountName,
  roleMemberCount,
}: {
  policy: AccessPolicy;
  memberEmail: (id: string) => string | undefined;
  serviceAccountName: (id: string) => string | undefined;
  roleMemberCount: (role: string) => number;
}) {
  const { type, value } = policy.subject;

  if (type === "all-members") {
    return (
      <Chip icon={<UsersIcon className={CHIP_ICON} strokeWidth={2} />}>
        All members
      </Chip>
    );
  }
  if (type === "role") {
    const n = roleMemberCount(value);
    return (
      <Chip icon={<IdentificationIcon className={CHIP_ICON} strokeWidth={2} />}>
        {value}
        {n > 0 && (
          <span className="text-text-muted font-normal ml-0.5">· {n}</span>
        )}
      </Chip>
    );
  }
  // A "user" subject is either a human member or a service account (bound by id).
  const sa = serviceAccountName(value);
  if (sa) {
    return (
      <Chip
        tone="primary"
        icon={<CpuChipIcon className={CHIP_ICON} strokeWidth={2} />}
      >
        {sa}
      </Chip>
    );
  }
  const email = memberEmail(value);
  return (
    <Chip
      icon={<UserIcon className={CHIP_ICON} strokeWidth={2} />}
      mono={!email}
      title={email ? undefined : value}
    >
      {email ?? `${value.slice(0, 12)}…`}
    </Chip>
  );
}

/* ── devices cell ─────────────────────────────────── */

function DevicesCell({ policy }: { policy: AccessPolicy }) {
  if (policy.filter.tags.length > 0) {
    // Keep rows compact: show up to two tags inline, collapse the rest into a "+N" chip.
    const tags = policy.filter.tags;
    const shown = tags.slice(0, 2);
    const rest = tags.slice(2);
    return (
      <span className="inline-flex items-center gap-1.5 flex-wrap">
        {shown.map((tag) => (
          <Chip
            key={tag.id}
            tone="primary"
            mono
            icon={<TagIcon className={CHIP_ICON} strokeWidth={2} />}
          >
            {tag.name}
          </Chip>
        ))}
        {rest.length > 0 && (
          <Chip title={rest.map((t) => t.name).join(", ")}>+{rest.length}</Chip>
        )}
      </span>
    );
  }
  if (policy.filter.hostname && policy.filter.hostname !== ".*") {
    return (
      <Chip mono icon={<ServerIcon className={CHIP_ICON} strokeWidth={2} />}>
        {policy.filter.hostname}
      </Chip>
    );
  }
  return (
    <Chip icon={<GlobeAltIcon className={CHIP_ICON} strokeWidth={2} />}>
      All devices
    </Chip>
  );
}

/* ── logins cell ──────────────────────────────────── */

function LoginsCell({ policy }: { policy: AccessPolicy }) {
  const isAny = policy.logins.length === 1 && policy.logins[0] === "*";
  if (isAny) {
    return (
      <Chip icon={<CommandLineIcon className={CHIP_ICON} strokeWidth={2} />}>
        Any login
      </Chip>
    );
  }
  return (
    <span className="inline-flex items-center gap-1.5 flex-wrap">
      {policy.logins.map((login) => (
        <Chip key={login} tone="cyan" mono>
          {login}
        </Chip>
      ))}
    </span>
  );
}

/* ── action cell ──────────────────────────────────── */

function ActionCell({ policy }: { policy: AccessPolicy }) {
  if (policy.action === "deny") {
    return (
      <Chip
        tone="red"
        icon={<NoSymbolIcon className={CHIP_ICON} strokeWidth={2} />}
      >
        Deny
      </Chip>
    );
  }
  return (
    <Chip
      tone="green"
      icon={<CheckCircleIcon className={CHIP_ICON} strokeWidth={2} />}
    >
      Allow
    </Chip>
  );
}

/* ── page ─────────────────────────────────────────── */

export default function AccessPolicies() {
  const { policies, isLoading } = useAccessPolicies();
  const { tenant: tenantId } = useAuthStore();
  const { namespace: ns } = useNamespace(tenantId ?? "");
  const { serviceAccounts } = useServiceAccounts();
  const isIdentityMode = ns?.settings?.ssh_access_mode === "identity";

  // Resolve a policy subject's id to a human label for the list. A "user" subject may be a
  // member (show email) or a service account (show name); a role shows its member count.
  const members = ns?.members ?? [];
  const memberEmail = (id: string) => members.find((m) => m.id === id)?.email;
  const serviceAccountName = (id: string) =>
    serviceAccounts.find((s) => s.id === id)?.name;
  const roleMemberCount = (role: string) =>
    members.filter((m) => String(m.role) === role).length;
  const deletePolicy = useDeleteAccessPolicy();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<AccessPolicy | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<AccessPolicy | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const [query, setQuery] = useState("");

  const filtered = policies.filter((p) =>
    p.name.toLowerCase().includes(query.trim().toLowerCase()),
  );

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
      key: "action",
      header: "Action",
      render: (p) => <ActionCell policy={p} />,
    },
    {
      key: "subject",
      header: "Subject",
      render: (p) => (
        <SubjectCell
          policy={p}
          memberEmail={memberEmail}
          serviceAccountName={serviceAccountName}
          roleMemberCount={roleMemberCount}
        />
      ),
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
              onClick={(e) => {
                e.stopPropagation();
                openEdit(p);
              }}
            >
              <PencilSquareIcon className="w-4 h-4" strokeWidth={2} />
            </IconButton>
          </RestrictedAction>
          <RestrictedAction action="accessPolicy:remove">
            <IconButton
              variant="danger"
              title="Delete"
              aria-label={`Delete ${p.name}`}
              onClick={(e) => {
                e.stopPropagation();
                setDeleteTarget(p);
              }}
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
              : "Decide who may reach which devices, as which login, under the identity SSH access mode. Policies are default-deny; deny rules win over allow."
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

      <div className="relative mb-3 max-w-xs">
        <MagnifyingGlassIcon
          className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted"
          strokeWidth={2}
        />
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search policies by name…"
          className="w-full pl-9 pr-3 py-2 bg-card border border-border rounded-lg text-sm text-text-primary placeholder:text-text-secondary outline-none focus:border-primary/60"
        />
      </div>

      <DataTable
        columns={columns}
        data={filtered}
        rowKey={(p) => p.id}
        isLoading={isLoading}
        loadingMessage="Loading access policies..."
        emptyMessage={
          query ? `No policies match "${query}"` : "No access policies found"
        }
        onRowClick={openEdit}
        rowClassName={() => "cursor-pointer"}
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
