import { useState } from "react";
import { useParams, useNavigate, Link } from "react-router-dom";
import {
  ChevronRightIcon,
  ServerStackIcon,
  PencilSquareIcon,
  TrashIcon,
  InformationCircleIcon,
  Cog6ToothIcon,
} from "@heroicons/react/24/outline";
import { useAdminNamespace } from "@/hooks/useAdminNamespaces";
import CopyButton from "@/components/common/CopyButton";
import DataTable, { type Column } from "@/components/common/DataTable";
import EditNamespaceDrawer from "./EditNamespaceDrawer";
import DeleteNamespaceDialog from "./DeleteNamespaceDialog";
import { formatDateFull } from "@/utils/date";
import { formatMaxDevices } from "./utils";

const LABEL
  = "text-2xs font-mono font-semibold uppercase tracking-label text-text-muted";
const VALUE = "text-sm text-text-primary font-medium mt-0.5";
const ZERO_DATE = "0001-01-01T00:00:00Z";

type Member = NonNullable<NonNullable<ReturnType<typeof useAdminNamespace>["data"]>["members"]>[number];

export default function NamespaceDetails() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data: namespace, isLoading, error } = useAdminNamespace(id ?? "");

  const [editOpen, setEditOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-24" role="status">
        <span className="w-5 h-5 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
        <span className="sr-only">Loading namespace details</span>
      </div>
    );
  }

  if (error || !namespace) {
    return (
      <div className="text-center py-24">
        <ServerStackIcon
          className="w-10 h-10 text-text-muted/30 mx-auto mb-3"
          strokeWidth={1}
        />
        <p className="text-sm text-text-muted mb-2">Namespace not found</p>
        <Link
          to="/admin/namespaces"
          className="text-sm text-primary hover:underline"
        >
          Back to namespaces
        </Link>
      </div>
    );
  }

  const ownerMember = namespace.members?.find((m) => m.id === namespace.owner);
  const ownerLabel = ownerMember?.email || namespace.owner;
  const totalDevices
    = (namespace.devices_accepted_count || 0)
      + (namespace.devices_pending_count || 0)
      + (namespace.devices_rejected_count || 0);

  const memberColumns: Column<Member>[] = [
    {
      key: "email",
      header: "Email",
      render: (member) =>
        member.id ? (
          <Link
            to={`/admin/users/${member.id}`}
            className="text-sm text-primary hover:underline"
          >
            {member.email || member.id}
          </Link>
        ) : (
          <span className="text-sm text-text-primary">
            {member.email || "\u2014"}
          </span>
        ),
    },
    {
      key: "role",
      header: "Role",
      render: (member) => (
        <span className="inline-flex items-center px-2 py-0.5 text-2xs font-semibold rounded-md bg-primary/10 text-primary border border-primary/20 capitalize">
          {member.role || "member"}
        </span>
      ),
    },
    {
      key: "added",
      header: "Added",
      render: (member) => (
        <span className="text-xs text-text-secondary">
          {member.added_at && member.added_at !== ZERO_DATE
            ? formatDateFull(member.added_at)
            : "\u2014"}
        </span>
      ),
    },
  ];

  return (
    <div className="animate-fade-in">
      {/* Breadcrumb */}
      <nav aria-label="Breadcrumb" className="flex items-center gap-1.5 mb-5">
        <Link
          to="/admin/namespaces"
          className="text-2xs font-mono text-text-muted hover:text-primary transition-colors"
        >
          Namespaces
        </Link>
        <ChevronRightIcon
          className="w-3 h-3 text-text-muted/40"
          strokeWidth={2}
        />
        <span className="text-2xs font-mono text-text-secondary">
          {namespace.name}
        </span>
      </nav>

      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-4 mb-8">
        <div className="flex items-start gap-4">
          <div className="w-14 h-14 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center shrink-0">
            <ServerStackIcon className="w-7 h-7 text-primary" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-text-primary">
              {namespace.name}
            </h1>
          </div>
        </div>

        <div className="flex items-center gap-2 shrink-0">
          <button
            onClick={() => setEditOpen(true)}
            className="flex items-center gap-2 px-4 py-2.5 border border-border text-text-secondary hover:text-text-primary hover:border-border-light rounded-lg text-sm font-semibold transition-all"
          >
            <PencilSquareIcon className="w-4 h-4" />
            Edit
          </button>
          <button
            onClick={() => setDeleteOpen(true)}
            className="p-2.5 rounded-lg text-text-muted hover:text-accent-red hover:bg-accent-red/10 border border-border transition-all"
            title="Delete namespace"
            aria-label={`Delete ${namespace.name}`}
          >
            <TrashIcon className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Info Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
        {/* Properties Card */}
        <div className="bg-card border border-border rounded-xl p-5 space-y-4">
          <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2">
            <InformationCircleIcon className="w-4 h-4 text-primary" />
            Properties
          </h3>
          <dl className="space-y-3">
            <div>
              <dt className={LABEL}>Name</dt>
              <dd className={VALUE}>{namespace.name}</dd>
            </div>
            <div>
              <dt className={LABEL}>Tenant ID</dt>
              <dd className="flex items-center gap-1 mt-0.5">
                <span className="text-xs font-mono text-text-primary">
                  {namespace.tenant_id}
                </span>
                <CopyButton text={namespace.tenant_id} />
              </dd>
            </div>
            <div>
              <dt className={LABEL}>Owner</dt>
              <dd className="mt-0.5">
                <Link
                  to={`/admin/users/${namespace.owner}`}
                  className="text-sm text-primary hover:underline"
                >
                  {ownerLabel}
                </Link>
              </dd>
            </div>
            <div>
              <dt className={LABEL}>Created</dt>
              <dd className={VALUE}>{formatDateFull(namespace.created_at)}</dd>
            </div>
          </dl>
        </div>

        {/* Settings Card */}
        <div className="bg-card border border-border rounded-xl p-5 space-y-4">
          <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2">
            <Cog6ToothIcon className="w-4 h-4 text-primary" />
            Settings
          </h3>
          <dl className="space-y-3">
            <div>
              <dt className={LABEL}>Max Devices</dt>
              <dd className={VALUE}>
                {formatMaxDevices(namespace.max_devices)}
              </dd>
            </div>
            <div>
              <dt className={LABEL}>Session Recording</dt>
              <dd className="mt-1">
                <span
                  className={`inline-flex items-center px-2 py-0.5 text-2xs font-semibold rounded-md ${
                    namespace.settings?.session_record
                      ? "bg-accent-green/10 text-accent-green border border-accent-green/20"
                      : "bg-accent-yellow/10 text-accent-yellow border border-accent-yellow/20"
                  }`}
                >
                  {namespace.settings?.session_record ? "Enabled" : "Disabled"}
                </span>
              </dd>
            </div>
            {namespace.settings?.connection_announcement && (
              <div>
                <dt className={LABEL}>Connection Announcement</dt>
                <dd className="mt-1.5 overflow-x-auto rounded-lg bg-surface border border-border p-3">
                  <pre className="text-xs font-mono text-text-primary whitespace-pre">
                    {namespace.settings.connection_announcement}
                  </pre>
                </dd>
              </div>
            )}
            <div>
              <dt className={LABEL}>Total Devices</dt>
              <dd className={VALUE}>{totalDevices}</dd>
            </div>
            <div className="flex items-center gap-6">
              <div>
                <dt className={LABEL}>Accepted</dt>
                <dd className={VALUE}>
                  {namespace.devices_accepted_count || 0}
                </dd>
              </div>
              <div>
                <dt className={LABEL}>Pending</dt>
                <dd className={VALUE}>
                  {namespace.devices_pending_count || 0}
                </dd>
              </div>
              <div>
                <dt className={LABEL}>Rejected</dt>
                <dd className={VALUE}>
                  {namespace.devices_rejected_count || 0}
                </dd>
              </div>
            </div>
          </dl>
        </div>
      </div>

      {/* Members Section */}
      <div className="bg-card border border-border rounded-xl overflow-hidden">
        <div className="px-5 py-4 border-b border-border">
          <h3 className="text-xs font-semibold text-text-primary">
            Members ({namespace.members?.length || 0})
          </h3>
        </div>
        <DataTable<Member>
          columns={memberColumns}
          data={namespace.members ?? []}
          rowKey={(m, i) => m.id || m.email || `member-${i}`}
          label="Members"
          noWrapper
          emptyMessage="No members"
        />
      </div>

      <EditNamespaceDrawer
        open={editOpen}
        onClose={() => setEditOpen(false)}
        namespace={namespace}
      />

      <DeleteNamespaceDialog
        open={deleteOpen}
        onClose={() => setDeleteOpen(false)}
        namespace={
          namespace
            ? { tenant_id: namespace.tenant_id, name: namespace.name }
            : null
        }
        onDeleted={() => void navigate("/admin/namespaces")}
      />
    </div>
  );
}
