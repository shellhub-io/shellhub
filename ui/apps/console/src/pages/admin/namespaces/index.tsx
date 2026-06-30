import { useState, type MouseEvent } from "react";
import { useNavigate } from "react-router-dom";
import {
  ServerStackIcon,
  PencilSquareIcon,
  TrashIcon,
} from "@heroicons/react/24/outline";
import { useAdminNamespaces } from "@/hooks/useAdminNamespaces";
import { useDebouncedValue } from "@/hooks/useDebouncedValue";
import { usePaginatedListState } from "@/hooks/usePaginatedListState";
import type { Namespace } from "@/client";
import PageHeader from "@/components/common/PageHeader";
import { DataTable, type Column } from "@shellhub/design-system/components";
import SearchField from "@/components/common/fields/SearchField";
import EditNamespaceDrawer from "./EditNamespaceDrawer";
import DeleteNamespaceDialog from "./DeleteNamespaceDialog";
import { formatDateShort } from "@/utils/date";
import { formatMaxDevices } from "./utils";
import { Callout, IconButton } from "@shellhub/design-system/primitives";

const PER_PAGE = 10;
const SEARCH_DEBOUNCE_MS = 300;

type AdminNamespacesParams = {
  page: number;
  search: string;
};

const DEFAULTS: AdminNamespacesParams = {
  page: 1,
  search: "",
};

function getOwnerEmail(namespace: Namespace): string {
  const owner = namespace.members?.find((m) => m.id === namespace.owner);
  return owner?.email || namespace.owner;
}

export default function AdminNamespaces() {
  const navigate = useNavigate();
  const { params, setPage, setSearch } =
    usePaginatedListState<AdminNamespacesParams>({ defaults: DEFAULTS });
  const debouncedSearch = useDebouncedValue(params.search, SEARCH_DEBOUNCE_MS);
  const [editTarget, setEditTarget] = useState<Namespace | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Namespace | null>(null);

  const { namespaces, totalCount, isLoading, error } = useAdminNamespaces({
    page: params.page,
    perPage: PER_PAGE,
    search: debouncedSearch,
  });

  const totalPages = Math.ceil(totalCount / PER_PAGE);

  const columns: Column<Namespace>[] = [
    {
      key: "name",
      header: "Name",
      render: (ns) => (
        <span className="text-sm font-medium text-text-primary group-hover:text-primary transition-colors">
          {ns.name}
        </span>
      ),
    },
    {
      key: "owner",
      header: "Owner",
      render: (ns) => (
        <span className="text-xs text-text-secondary">{getOwnerEmail(ns)}</span>
      ),
    },
    {
      key: "devices",
      header: "Devices",
      render: (ns) => (
        <span className="text-xs text-text-secondary">
          {ns.devices_accepted_count}
        </span>
      ),
    },
    {
      key: "max_devices",
      header: "Max Devices",
      render: (ns) => (
        <span className="text-xs text-text-secondary">
          {formatMaxDevices(ns.max_devices)}
        </span>
      ),
    },
    {
      key: "created",
      header: "Created",
      render: (ns) => (
        <span className="text-xs text-text-secondary">
          {formatDateShort(ns.created_at)}
        </span>
      ),
    },
    {
      key: "actions",
      header: "Actions",
      headerClassName: "text-right",
      render: (ns) => (
        <div className="flex items-center justify-end gap-1">
          <IconButton
            variant="primary"
            title="Edit namespace"
            aria-label={`Edit ${ns.name}`}
            onClick={(e: MouseEvent) => {
              e.stopPropagation();
              setEditTarget(ns);
            }}
          >
            <PencilSquareIcon className="w-4 h-4" />
          </IconButton>
          <IconButton
            variant="danger"
            title="Delete namespace"
            aria-label={`Delete ${ns.name}`}
            onClick={(e: MouseEvent) => {
              e.stopPropagation();
              setDeleteTarget(ns);
            }}
          >
            <TrashIcon className="w-4 h-4" />
          </IconButton>
        </div>
      ),
    },
  ];

  return (
    <div>
      <PageHeader
        icon={<ServerStackIcon className="w-6 h-6" />}
        overline="Namespace Management"
        title="Namespaces"
        description="Manage all namespaces in the instance"
      />

      <SearchField
        className="mb-5"
        value={params.search}
        onChange={setSearch}
        placeholder="Search by name..."
        aria-label="Search namespaces by name"
      />

      {error && (
        <Callout variant="error" className="mb-4">
          {error.message}
        </Callout>
      )}

      <DataTable
        columns={columns}
        data={namespaces}
        rowKey={(ns) => ns.tenant_id}
        isLoading={isLoading}
        loadingMessage="Loading namespaces..."
        page={params.page}
        totalPages={totalPages}
        totalCount={totalCount}
        itemLabel="namespace"
        onPageChange={setPage}
        onRowClick={(ns) => void navigate(`/admin/namespaces/${ns.tenant_id}`)}
        emptyState={
          <div className="text-center">
            <ServerStackIcon
              className="w-10 h-10 text-text-muted/30 mx-auto mb-3"
              strokeWidth={1}
            />
            <p className="text-xs font-mono text-text-muted">
              {debouncedSearch
                ? `No namespaces matching "${debouncedSearch}"`
                : "No namespaces found"}
            </p>
          </div>
        }
      />

      <EditNamespaceDrawer
        open={!!editTarget}
        onClose={() => setEditTarget(null)}
        namespace={editTarget}
      />

      <DeleteNamespaceDialog
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        namespace={deleteTarget}
      />
    </div>
  );
}
