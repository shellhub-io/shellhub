import { useNavigate, Link } from "react-router-dom";
import {
  CpuChipIcon,
} from "@heroicons/react/24/outline";
import { Callout } from "@shellhub/design-system/primitives";
import {
  useAdminDevices,
  type NormalizedDevice,
} from "@/hooks/useAdminDevices";
import type { DeviceStatus } from "@/client";
import PageHeader from "@/components/common/PageHeader";
import { DataTable, type Column } from "@shellhub/design-system/components";
import SearchField from "@/components/common/fields/SearchField";
import DistroIcon from "@/components/common/DistroIcon";
import { useDebouncedValue } from "@/hooks/useDebouncedValue";
import { usePaginatedListState } from "@/hooks/usePaginatedListState";
import DeviceStatusChip from "./DeviceStatusChip";
import { formatRelative } from "@/utils/date";

const PER_PAGE = 10;
const SEARCH_DEBOUNCE_MS = 300;

type StatusTab = { label: string; value: DeviceStatus | "" };

const statusTabs: StatusTab[] = [
  { label: "All", value: "" },
  { label: "Accepted", value: "accepted" },
  { label: "Pending", value: "pending" },
  { label: "Rejected", value: "rejected" },
];

const VALID_STATUSES = ["", "accepted", "pending", "rejected"] as const;
const VALID_SORT_FIELDS = ["name", "last_seen", "status"] as const;
const VALID_SORT_ORDERS = ["asc", "desc"] as const;

/** Stable module-level constants — avoids new object/array identities every
 *  render, which would invalidate the `update` and `handleSort` useCallbacks
 *  in usePaginatedListState and cascade unnecessary re-renders to children. */
const CONSTRAINTS = {
  status: VALID_STATUSES,
  sortField: VALID_SORT_FIELDS,
  sortOrder: VALID_SORT_ORDERS,
} as const;

const SORT_FIELDS = [
  { field: "name", initialOrder: "asc" as const },
  { field: "last_seen", initialOrder: "desc" as const },
  { field: "status", initialOrder: "desc" as const },
];

type SortField = typeof VALID_SORT_FIELDS[number];

type AdminDevicesParams = {
  page: number;
  search: string;
  sortField: SortField;
  sortOrder: "asc" | "desc";
  status: DeviceStatus | "";
};

const DEFAULTS: AdminDevicesParams = {
  page: 1,
  search: "",
  sortField: "last_seen",
  sortOrder: "desc",
  status: "",
};

function TagChips({ tags }: { tags: string[] }) {
  if (tags.length === 0) {
    return (
      <span className="text-2xs text-text-muted/50 font-mono">No tags</span>
    );
  }
  return (
    <div className="flex items-center gap-1 flex-wrap">
      {tags.map((tag, i) => (
        <span
          key={`${tag}-${i}`}
          className="inline-flex items-center px-1.5 py-0.5 bg-primary/8 text-primary text-2xs rounded font-medium border border-primary/15"
          title={tag}
        >
          {tag.length > 10 ? `${tag.slice(0, 10)}...` : tag}
        </span>
      ))}
    </div>
  );
}

export default function AdminDevices() {
  const navigate = useNavigate();

  const { params, setPage, setSearch, setFilter, handleSort } =
    usePaginatedListState<AdminDevicesParams>({
      defaults: DEFAULTS,
      constraints: CONSTRAINTS,
      sortFields: SORT_FIELDS,
    });

  const debouncedSearch = useDebouncedValue(params.search, SEARCH_DEBOUNCE_MS);

  const { devices, totalCount, isLoading, error } = useAdminDevices({
    page: params.page,
    perPage: PER_PAGE,
    search: debouncedSearch,
    status: params.status,
    sortBy: params.sortField,
    orderBy: params.sortOrder,
  });

  const totalPages = Math.ceil(totalCount / PER_PAGE);

  const columns: Column<NormalizedDevice>[] = [
    {
      key: "online",
      header: "",
      headerClassName: "w-12",
      render: (device) =>
        device.online ? (
          <span className="relative flex h-2.5 w-2.5 mx-auto" title="Online">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent-green opacity-40" />
            <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-accent-green shadow-[0_0_6px_rgba(130,165,104,0.4)]" />
          </span>
        ) : (
          <span
            className="block w-2.5 h-2.5 rounded-full mx-auto bg-text-muted/30"
            title="Offline"
          />
        ),
    },
    {
      key: "name",
      header: "Hostname",
      sortable: true,
      render: (device) => (
        <span className="text-sm font-medium text-text-primary group-hover:text-primary transition-colors truncate block max-w-[200px]">
          {device.name}
        </span>
      ),
    },
    {
      key: "os",
      header: "Operating System",
      render: (device) => (
        <div className="flex items-center gap-2">
          <DistroIcon
            id={device.info?.id ?? ""}
            className="text-base leading-none"
          />
          <span className="text-xs text-text-secondary truncate max-w-[160px]">
            {device.info?.pretty_name ?? "Unknown"}
          </span>
        </div>
      ),
    },
    {
      key: "namespace",
      header: "Namespace",
      render: (device) =>
        device.namespace ? (
          <Link
            to={`/admin/namespaces/${device.tenant_id}`}
            onClick={(e) => e.stopPropagation()}
            className="text-xs text-primary hover:underline"
          >
            {device.namespace}
          </Link>
        ) : (
          <span className="text-xs text-text-muted">&mdash;</span>
        ),
    },
    {
      key: "tags",
      header: "Tags",
      render: (device) => <TagChips tags={device.tags} />,
    },
    {
      key: "last_seen",
      header: "Last Seen",
      sortable: true,
      render: (device) => (
        <span className="text-xs text-text-secondary">
          {formatRelative(device.last_seen)}
        </span>
      ),
    },
    {
      key: "status",
      header: "Status",
      sortable: true,
      render: (device) => <DeviceStatusChip status={device.status} />,
    },
  ];

  return (
    <div>
      <PageHeader
        icon={<CpuChipIcon className="w-6 h-6" />}
        overline="Device Administration"
        title="Devices"
        description="View all devices registered across the instance"
      />

      {/* Filter bar */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 mb-5 animate-fade-in">
        <div
          className="flex items-center h-8 bg-card border border-border rounded-md p-0.5"
          role="tablist"
          aria-label="Filter devices by status"
        >
          {statusTabs.map((tab) => (
            <button
              type="button"
              key={tab.value}
              role="tab"
              aria-selected={params.status === tab.value}
              onClick={() => setFilter("status", tab.value)}
              className={`h-full px-3.5 text-xs font-medium rounded transition-all duration-150 ${
                params.status === tab.value
                  ? "bg-primary/15 text-primary border border-primary/25"
                  : "text-text-muted hover:text-text-secondary border border-transparent"
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>

        <SearchField
          value={params.search}
          onChange={(next) => setSearch(next)}
          placeholder="Search by hostname..."
          aria-label="Search devices by hostname"
        />
      </div>

      {error && (
        <Callout variant="error" className="mb-4">
          {error.message}
        </Callout>
      )}

      <DataTable
        columns={columns}
        data={devices}
        rowKey={(device) => device.uid}
        isLoading={isLoading}
        loadingMessage="Loading devices..."
        page={params.page}
        totalPages={totalPages}
        totalCount={totalCount}
        itemLabel="device"
        onPageChange={setPage}
        onRowClick={(device) => void navigate(`/admin/devices/${device.uid}`)}
        sortField={params.sortField}
        sortOrder={params.sortOrder}
        onSort={handleSort}
        emptyState={
          <div className="text-center">
            <CpuChipIcon
              className="w-10 h-10 text-text-muted/30 mx-auto mb-3"
              strokeWidth={1}
            />
            <p className="text-xs font-mono text-text-muted">
              {debouncedSearch
                ? `No devices matching "${debouncedSearch}"`
                : "No devices found"}
            </p>
          </div>
        }
      />
    </div>
  );
}
