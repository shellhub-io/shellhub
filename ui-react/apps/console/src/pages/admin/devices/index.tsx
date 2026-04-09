import { useState, useEffect } from "react";
import { useNavigate, Link } from "react-router-dom";
import {
  CpuChipIcon,
  MagnifyingGlassIcon,
  ExclamationCircleIcon,
} from "@heroicons/react/24/outline";
import {
  useAdminDevices,
  type NormalizedDevice,
} from "@/hooks/useAdminDevices";
import type { DeviceStatus } from "@/client";
import PageHeader from "@/components/common/PageHeader";
import DataTable, { type Column } from "@/components/common/DataTable";
import DistroIcon from "@/components/common/DistroIcon";
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

type SortField = "name" | "last_seen" | "status";

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
  const [page, setPage] = useState(1);
  const [searchInput, setSearchInput] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [status, setStatus] = useState<DeviceStatus | "">("");
  const [sortBy, setSortBy] = useState<SortField>("last_seen");
  const [orderBy, setOrderBy] = useState<"asc" | "desc">("desc");

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchInput);
      setPage(1);
    }, SEARCH_DEBOUNCE_MS);
    return () => clearTimeout(timer);
  }, [searchInput]);

  const { devices, totalCount, isLoading, error } = useAdminDevices({
    page,
    perPage: PER_PAGE,
    search: debouncedSearch,
    status,
    sortBy,
    orderBy,
  });

  const totalPages = Math.ceil(totalCount / PER_PAGE);

  const handleStatusChange = (newStatus: DeviceStatus | "") => {
    setStatus(newStatus);
    setPage(1);
  };

  const handleSort = (field: string) => {
    const f = field as SortField;
    if (sortBy === f) {
      setOrderBy((prev) => (prev === "asc" ? "desc" : "asc"));
    } else {
      setSortBy(f);
      setOrderBy(f === "name" ? "asc" : "desc");
    }
    setPage(1);
  };

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
              key={tab.value}
              role="tab"
              aria-selected={status === tab.value}
              onClick={() => handleStatusChange(tab.value)}
              className={`h-full px-3.5 text-xs font-medium rounded transition-all duration-150 ${
                status === tab.value
                  ? "bg-primary/15 text-primary border border-primary/25"
                  : "text-text-muted hover:text-text-secondary border border-transparent"
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>

        <div className="relative h-8">
          <MagnifyingGlassIcon
            className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-text-muted"
            strokeWidth={2}
          />
          <input
            type="text"
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
            placeholder="Search by hostname..."
            aria-label="Search devices by hostname"
            className="h-full pl-9 pr-3 bg-card border border-border rounded-md text-xs text-text-primary font-mono placeholder:text-text-secondary focus:outline-none focus:border-primary/40 focus:ring-1 focus:ring-primary/15 transition-all duration-200 w-56"
          />
        </div>
      </div>

      {error && (
        <div
          role="alert"
          className="flex items-center gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3.5 py-2.5 rounded-md text-xs font-mono mb-4 animate-slide-down"
        >
          <ExclamationCircleIcon
            className="w-3.5 h-3.5 shrink-0"
            strokeWidth={2}
          />
          {error.message}
        </div>
      )}

      <DataTable
        columns={columns}
        data={devices}
        rowKey={(device) => device.uid}
        isLoading={isLoading}
        loadingMessage="Loading devices..."
        page={page}
        totalPages={totalPages}
        totalCount={totalCount}
        itemLabel="device"
        onPageChange={setPage}
        onRowClick={(device) => void navigate(`/admin/devices/${device.uid}`)}
        sortField={sortBy}
        sortOrder={orderBy}
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
