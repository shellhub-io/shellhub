import { useState } from "react";
import { useParams, useNavigate, Link } from "react-router-dom";
import {
  ServerStackIcon,
  PencilSquareIcon,
  TrashIcon,
  InformationCircleIcon,
  Cog6ToothIcon,
} from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import { useAdminNamespace } from "@/hooks/useAdminNamespaces";
import Breadcrumb from "@/components/common/Breadcrumb";
import DataTable, { type Column } from "@/components/common/DataTable";
import EditNamespaceDrawer from "./EditNamespaceDrawer";
import DeleteNamespaceDialog from "./DeleteNamespaceDialog";
import { formatDateFull } from "@/utils/date";
import { formatMaxDevices } from "./utils";
import InfoItem from "@/components/common/InfoItem";
import PageLoader from "@/components/common/PageLoader";
import {
  Badge,
  Button,
  Card,
  IconButton,
} from "@shellhub/design-system/primitives";

const ZERO_DATE = "0001-01-01T00:00:00Z";

type Member = NonNullable<
  NonNullable<ReturnType<typeof useAdminNamespace>["data"]>["members"]
>[number];

export default function NamespaceDetails() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data: namespace, isLoading, error } = useAdminNamespace(id ?? "");

  const [editOpen, setEditOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);

  if (isLoading) {
    return <PageLoader label="Loading namespace details" />;
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
  const totalDevices =
    (namespace.devices_accepted_count || 0) +
    (namespace.devices_pending_count || 0) +
    (namespace.devices_rejected_count || 0);

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
        <Badge color="primary" className="capitalize">
          {member.role || "member"}
        </Badge>
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
      <Breadcrumb
        items={[
          { label: "Namespaces", to: "/admin/namespaces" },
          { label: namespace.name },
        ]}
      />

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
          <Button
            variant="secondary"
            icon={<PencilSquareIcon className="w-4 h-4" />}
            onClick={() => setEditOpen(true)}
          >
            Edit
          </Button>
          <IconButton
            variant="danger"
            size="lg"
            title="Delete namespace"
            aria-label={`Delete ${namespace.name}`}
            className="border border-border"
            onClick={() => setDeleteOpen(true)}
          >
            <TrashIcon className="w-4 h-4" />
          </IconButton>
        </div>
      </div>

      {/* Info Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
        {/* Properties Card */}
        <Card className="p-5 space-y-4">
          <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2">
            <InformationCircleIcon className="w-4 h-4 text-primary" />
            Properties
          </h3>
          <dl className="space-y-3">
            <InfoItem label="Name" value={namespace.name} />
            <InfoItem
              label="Tenant ID"
              value={namespace.tenant_id}
              mono
              copyable
            />
            <InfoItem label="Owner">
              <Link
                to={`/admin/users/${namespace.owner}`}
                className="text-sm text-primary hover:underline"
              >
                {ownerLabel}
              </Link>
            </InfoItem>
            <InfoItem
              label="Created"
              value={formatDateFull(namespace.created_at)}
            />
          </dl>
        </Card>

        {/* Settings Card */}
        <Card className="p-5 space-y-4">
          <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2">
            <Cog6ToothIcon className="w-4 h-4 text-primary" />
            Settings
          </h3>
          <dl className="space-y-3">
            <InfoItem
              label="Max Devices"
              value={formatMaxDevices(namespace.max_devices)}
            />
            <InfoItem label="Session Recording">
              <span
                className={cn(
                  "inline-flex items-center px-2 py-0.5 text-2xs font-semibold rounded-md",
                  namespace.settings?.session_record
                    ? "bg-accent-green/10 text-accent-green border border-accent-green/20"
                    : "bg-accent-yellow/10 text-accent-yellow border border-accent-yellow/20",
                )}
              >
                {namespace.settings?.session_record ? "Enabled" : "Disabled"}
              </span>
            </InfoItem>
            {namespace.settings?.connection_announcement && (
              <InfoItem label="Connection Announcement">
                <div className="overflow-x-auto rounded-lg bg-surface border border-border p-3">
                  <pre className="text-xs font-mono text-text-primary whitespace-pre">
                    {namespace.settings.connection_announcement}
                  </pre>
                </div>
              </InfoItem>
            )}
            <div className="flex items-center gap-6">
              <InfoItem label="Total Devices" value={String(totalDevices)} />
              <InfoItem
                label="Accepted"
                value={String(namespace.devices_accepted_count || 0)}
              />
              <InfoItem
                label="Pending"
                value={String(namespace.devices_pending_count || 0)}
              />
              <InfoItem
                label="Rejected"
                value={String(namespace.devices_rejected_count || 0)}
              />
            </div>
          </dl>
        </Card>
      </div>

      {/* Members Section */}
      <Card className="overflow-hidden">
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
      </Card>

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
