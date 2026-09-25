import { useState, type ReactNode } from "react";
import {
  ShieldCheckIcon,
  PlusIcon,
  TagIcon,
  UsersIcon,
  UserIcon,
  KeyIcon,
  GlobeAltIcon,
  CommandLineIcon,
  PencilSquareIcon,
  TrashIcon,
  EllipsisVerticalIcon,
  IdentificationIcon,
  ExclamationTriangleIcon,
  CheckCircleIcon,
  NoSymbolIcon,
} from "@heroicons/react/24/outline";
import {
  Button,
  Dropdown,
  IconButton,
} from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { useAccessPolicies } from "@/hooks/useAccessPolicies";
import { useDeleteAccessPolicy } from "@/hooks/useAccessPolicyMutations";
import { useNamespace } from "@/hooks/useNamespaces";
import { useApiKeys } from "@/hooks/useApiKeys";
import { useAuthStore } from "@/stores/authStore";
import type { AccessPolicy } from "@/client";
import { roleSubjectCount as countRoleSubject } from "./subjectCount";
import PageHeader from "@/components/common/PageHeader";
import EmptyState from "@/components/common/EmptyState";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import DataTable, { type Column } from "@/components/common/DataTable";
import RestrictedAction from "@/components/common/RestrictedAction";
import SearchField from "@/components/common/fields/SearchField";
import { formatRelative } from "@/utils/date";
import { useNavSectionTitle } from "@/components/layout/navSections";
import AccessPolicyModal from "./AccessPolicyModal";

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

function SubjectCell({
  policy,
  memberEmail,
  apiKeyName,
  roleSubjectCount,
}: {
  policy: AccessPolicy;
  memberEmail: (id: string) => string | undefined;
  apiKeyName: (id: string) => string | undefined;
  roleSubjectCount: (role: string) => number;
}) {
  const { type, value } = policy.subject;

  if (!policy.subject_matches) {
    return (
      <Chip
        tone="red"
        icon={<ExclamationTriangleIcon className={CHIP_ICON} strokeWidth={2} />}
        title={
          policy.action === "deny"
            ? "Nobody this subject names can connect, so the rule blocks nobody."
            : "Nobody this subject names can connect, so the rule grants nothing."
        }
      >
        {type === "all-members" ? "All members" : value}
      </Chip>
    );
  }

  if (type === "all-members") {
    return (
      <Chip icon={<UsersIcon className={CHIP_ICON} strokeWidth={2} />}>
        All members
      </Chip>
    );
  }
  if (type === "role") {
    const n = roleSubjectCount(value);
    return (
      <Chip icon={<IdentificationIcon className={CHIP_ICON} strokeWidth={2} />}>
        {value}
        {n > 0 && (
          <span className="text-text-muted font-normal ml-0.5">· {n}</span>
        )}
      </Chip>
    );
  }
  if (type === "api-key") {
    return (
      <Chip
        tone="primary"
        icon={<KeyIcon className={CHIP_ICON} strokeWidth={2} />}
        title={apiKeyName(value) ? undefined : value}
      >
        {apiKeyName(value) ?? `${value.slice(0, 12)}…`}
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

function DevicesCell({ policy }: { policy: AccessPolicy }) {
  if (policy.filter.tags.length > 0) {
    const tags = policy.filter.tags;
    const shown = tags.slice(0, 2);
    const rest = tags.slice(2);
    return (
      <span className="inline-flex items-center gap-1.5 flex-wrap">
        {shown.map((tag) => (
          <Chip
            key={tag.name}
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
  return (
    <Chip icon={<GlobeAltIcon className={CHIP_ICON} strokeWidth={2} />}>
      All devices
    </Chip>
  );
}

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

/**
 * The access policies page: who may reach which devices, as what. This replaces the key ACL and
 * firewall pages in namespaces using identity access mode.
 */
export default function AccessPolicies() {
  const sectionTitle = useNavSectionTitle("/access-policies");
  const { policies, isLoading } = useAccessPolicies();
  const { tenant: tenantId } = useAuthStore();
  const { namespace: ns } = useNamespace(tenantId ?? "");
  const isIdentityMode = ns?.settings?.ssh_access_mode === "identity";

  const { apiKeys } = useApiKeys({ perPage: 100 });
  const members = ns?.members ?? [];
  const memberEmail = (id: string) => members.find((m) => m.id === id)?.email;
  const apiKeyName = (id: string) => apiKeys.find((k) => k.id === id)?.name;
  const roleSubjectCount = (role: string) =>
    countRoleSubject({ role, members });
  const deletePolicy = useDeleteAccessPolicy();
  const [modalOpen, setModalOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<AccessPolicy | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<AccessPolicy | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const [search, setSearch] = useState("");

  const filtered = policies.filter((p) =>
    p.name.toLowerCase().includes(search.trim().toLowerCase()),
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
    setModalOpen(true);
  };
  const openEdit = (policy: AccessPolicy) => {
    setEditTarget(policy);
    setModalOpen(true);
  };
  const closeModal = () => {
    setModalOpen(false);
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
          apiKeyName={apiKeyName}
          roleSubjectCount={roleSubjectCount}
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
          {formatRelative(p.created_at)}
        </span>
      ),
    },
    {
      key: "actions",
      header: "",
      headerClassName: "w-16",
      render: (p) => (
        <div
          className="flex justify-end"
          onClick={(e) => e.stopPropagation()}
          role="presentation"
        >
          <Dropdown portal placement="bottom-end">
            <Dropdown.Trigger>
              <IconButton variant="ghost" aria-label={`Actions for ${p.name}`}>
                <EllipsisVerticalIcon className="w-4 h-4" />
              </IconButton>
            </Dropdown.Trigger>

            <Dropdown.Panel className="w-40 py-1">
              <RestrictedAction action="accessPolicy:edit">
                <Dropdown.Item
                  label="Edit"
                  onSelect={() => openEdit(p)}
                  className="gap-2.5 px-3 py-2"
                >
                  <PencilSquareIcon className="w-4 h-4" />
                  Edit
                </Dropdown.Item>
              </RestrictedAction>
              <RestrictedAction action="accessPolicy:remove">
                <Dropdown.Item
                  label="Delete"
                  variant="danger"
                  onSelect={() => setDeleteTarget(p)}
                  className="gap-2.5 px-3 py-2"
                >
                  <TrashIcon className="w-4 h-4" />
                  Delete
                </Dropdown.Item>
              </RestrictedAction>
            </Dropdown.Panel>
          </Dropdown>
        </div>
      ),
    },
  ];

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
              description: "Target all devices, or scope to specific tags.",
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

        <AccessPolicyModal
          open={modalOpen}
          editPolicy={editTarget}
          onClose={closeModal}
        />
      </>
    );
  }

  return (
    <div>
      <PageHeader
        icon={<ShieldCheckIcon className="w-6 h-6" />}
        overline={sectionTitle}
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

      <SearchField
        className="mb-3"
        value={search}
        onChange={setSearch}
        placeholder="Search policies by name…"
        aria-label="Search access policies by name"
      />

      <DataTable
        columns={columns}
        data={filtered}
        rowKey={(p) => p.id}
        isLoading={isLoading}
        loadingMessage="Loading access policies..."
        emptyMessage={
          search ? `No policies match "${search}"` : "No access policies found"
        }
        onRowClick={openEdit}
        rowClassName={() => "cursor-pointer"}
      />

      <AccessPolicyModal
        open={modalOpen}
        editPolicy={editTarget}
        onClose={closeModal}
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
