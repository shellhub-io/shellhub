import { useState, useMemo, useCallback } from "react";
import { useNavigate, Link } from "react-router-dom";
import { useDevices, type NormalizedDevice } from "@/hooks/useDevices";
import { useDebouncedValue } from "@/hooks/useDebouncedValue";
import { useTableSort } from "@/hooks/useTableSort";
import { usePaginatedListState } from "@/hooks/usePaginatedListState";
import { useNamespace } from "@/hooks/useNamespaces";
import { useAuthStore } from "@/stores/authStore";
import { useTerminalStore } from "@/stores/terminalStore";
import { useDeviceActions } from "@/hooks/useDeviceActions";
import PageHeader from "@/components/common/PageHeader";
import ConnectDrawer from "@/components/ConnectDrawer";
import ManageTagsDrawer from "@/components/ManageTagsDrawer";
import CopyButton from "@/components/common/CopyButton";
import PlatformBadge from "@/components/common/PlatformBadge";
import OnlineDot from "@/components/common/OnlineDot";
import LastSeenCell from "@/components/common/LastSeenCell";
import { DataTable, type Column } from "@shellhub/design-system/components";
import SearchField from "@/components/common/fields/SearchField";
import { buildSshid } from "@/utils/sshid";
import TagFilterDropdown from "@/components/common/TagFilterDropdown";
import TagsPopover from "./TagsPopover";
import DeviceActionsPortal from "./DeviceActionsPortal";
import {
  PlusIcon,
  TagIcon,
  XMarkIcon,
  CpuChipIcon,
  ChevronDoubleRightIcon,
} from "@heroicons/react/24/outline";
import {
  Button,
  Callout,
  IconButton,
} from "@shellhub/design-system/primitives";
import RestrictedAction from "@/components/common/RestrictedAction";

const PER_PAGE = 10;
const SEARCH_DEBOUNCE_MS = 300;

const VALID_STATUSES = ["accepted", "pending", "rejected"] as const;

/** Stable module-level constant — avoids a new object identity every render,
 *  which would invalidate the `update` useCallback in usePaginatedListState. */
const CONSTRAINTS = { status: VALID_STATUSES } as const;

type ValidStatus = (typeof VALID_STATUSES)[number];

const statusTabs: { label: string; value: ValidStatus }[] = [
  { label: "Accepted", value: "accepted" },
  { label: "Pending", value: "pending" },
  { label: "Rejected", value: "rejected" },
];

type DevicesParams = {
  page: number;
  search: string;
  status: ValidStatus;
  tags: string[];
};

const DEFAULTS: DevicesParams = {
  page: 1,
  search: "",
  status: "accepted",
  tags: [],
};

type SortField = "name" | "last_seen";

export default function Devices() {
  const { params, setPage, setSearch, setFilter, setArrayFilter, mapArrayFilter } =
    usePaginatedListState<DevicesParams>({
      defaults: DEFAULTS,
      constraints: CONSTRAINTS,
    });

  const debouncedSearch = useDebouncedValue(
    params.search.trim(),
    SEARCH_DEBOUNCE_MS,
  );

  const deviceActions = useDeviceActions();
  const { requestAction: requestDeviceAction } = deviceActions;
  const [connectTarget, setConnectTarget] = useState<{
    uid: string;
    name: string;
    sshid: string;
  } | null>(null);
  const [manageTagsOpen, setManageTagsOpen] = useState(false);
  const { sortBy, orderBy, handleSort } = useTableSort<SortField>({
    defaultField: "last_seen",
    onSortChange: () => setPage(1),
  });

  const { devices, totalCount, isLoading, error, refetch } = useDevices({
    page: params.page,
    perPage: PER_PAGE,
    status: params.status,
    search: debouncedSearch,
    filterTags: params.tags,
    sortBy,
    orderBy,
  });

  const tenantId = useAuthStore((s) => s.tenant) ?? "";
  const { namespace: currentNamespace } = useNamespace(tenantId);
  const navigate = useNavigate();

  const totalPages = Math.ceil(totalCount / PER_PAGE);
  const nsName = currentNamespace?.name ?? "";

  const handleStatusChange = (newStatus: ValidStatus) => {
    setFilter("status", newStatus);
  };

  const addFilterTag = useCallback(
    (tag: string) => {
      mapArrayFilter("tags", (tags) =>
        tags.includes(tag) ? tags : [...tags, tag],
      );
    },
    [mapArrayFilter],
  );

  const removeFilterTag = (tag: string) => {
    mapArrayFilter("tags", (tags) => tags.filter((t) => t !== tag));
  };

  const clearFilterTags = () => {
    setArrayFilter("tags", []);
  };

  const columns = useMemo<Column<NormalizedDevice>[]>(() => {
    const baseColumns: Column<NormalizedDevice>[] = [
      {
        key: "name",
        header: "Hostname",
        sortable: true,
        render: (device) => (
          <span className="text-sm font-medium text-text-primary group-hover:text-primary transition-colors">
            {device.name}
          </span>
        ),
      },
      {
        key: "os",
        header: "Operating System",
        render: (device) => (
          <div className="flex items-center gap-2">
            <span className="text-xs text-text-secondary truncate max-w-[160px]">
              {device.info?.pretty_name ?? "Unknown"}
            </span>
            {device.info?.platform && (
              <PlatformBadge platform={device.info.platform} />
            )}
          </div>
        ),
      },
      {
        key: "tags",
        header: "Tags",
        render: (device) => (
          <TagsPopover device={device} onFilterTag={addFilterTag} />
        ),
      },
      {
        key: "last_seen",
        header: "Last Seen",
        sortable: true,
        render: (device) => <LastSeenCell value={device.last_seen} />,
      },
    ];

    if (params.status === "accepted") {
      return [
        {
          key: "online",
          header: "",
          headerClassName: "w-12",
          render: (device) => <OnlineDot online={device.online} />,
        },
        baseColumns[0], // hostname
        {
          key: "sshid",
          header: "SSHID",
          render: (device) => {
            const sshid = nsName
              ? buildSshid(nsName, device.name)
              : device.uid.substring(0, 8);
            return (
              <div className="flex items-center gap-1">
                <code
                  className="text-2xs font-mono text-text-muted truncate max-w-[220px]"
                  title={sshid}
                >
                  {sshid}
                </code>
                <CopyButton text={sshid} />
              </div>
            );
          },
        },
        ...baseColumns.slice(1), // os, tags, last_seen
        {
          key: "connect",
          header: "",
          headerClassName: "w-20",
          render: (device) =>
            device.online ? (
              <RestrictedAction action="device:connect">
                <Button
                  variant="successSoft"
                  size="sm"
                  icon={
                    <ChevronDoubleRightIcon
                      className="w-3 h-3"
                      strokeWidth={2}
                    />
                  }
                  onClick={(e) => {
                    e.stopPropagation();
                    const existing = useTerminalStore
                      .getState()
                      .sessions.find((s) => s.deviceUid === device.uid);
                    if (existing) {
                      useTerminalStore.getState().restore(existing.id);
                    } else {
                      const sshid = nsName
                        ? buildSshid(nsName, device.name)
                        : device.uid;
                      setConnectTarget({
                        uid: device.uid,
                        name: device.name,
                        sshid,
                      });
                    }
                  }}
                >
                  Connect
                </Button>
              </RestrictedAction>
            ) : (
              <span className="text-2xs text-text-muted/30 font-mono">
                Offline
              </span>
            ),
        },
      ];
    }

    if (params.status === "pending") {
      return [
        ...baseColumns,
        {
          key: "actions",
          header: "Actions",
          headerClassName: "text-right",
          render: (device) => (
            <div className="flex items-center justify-end gap-1.5">
              <RestrictedAction action="device:accept">
                <Button
                  variant="successSoft"
                  size="sm"
                  onClick={(e) => {
                    e.stopPropagation();
                    requestDeviceAction(device, "accept");
                  }}
                >
                  Accept
                </Button>
              </RestrictedAction>
              <RestrictedAction action="device:reject">
                <Button
                  variant="warningSoft"
                  size="sm"
                  onClick={(e) => {
                    e.stopPropagation();
                    requestDeviceAction(device, "reject");
                  }}
                >
                  Reject
                </Button>
              </RestrictedAction>
            </div>
          ),
        },
      ];
    }

    // rejected
    return [
      ...baseColumns,
      {
        key: "actions",
        header: "Actions",
        headerClassName: "text-right",
        render: (device) => (
          <div className="flex items-center justify-end gap-1.5">
            <RestrictedAction action="device:accept">
              <Button
                variant="successSoft"
                size="sm"
                onClick={(e) => {
                  e.stopPropagation();
                  requestDeviceAction(device, "accept");
                }}
              >
                Accept
              </Button>
            </RestrictedAction>
            <RestrictedAction action="device:remove">
              <Button
                variant="dangerSoft"
                size="sm"
                onClick={(e) => {
                  e.stopPropagation();
                  requestDeviceAction(device, "remove");
                }}
              >
                Remove
              </Button>
            </RestrictedAction>
          </div>
        ),
      },
    ];
  }, [params.status, nsName, addFilterTag, requestDeviceAction]);

  return (
    <div>
      <PageHeader
        icon={<CpuChipIcon className="w-6 h-6" />}
        overline="Device Management"
        title="Devices"
        description="Manage and monitor all devices connected to your namespace"
      >
        <RestrictedAction action="device:add">
          <Button
            as={Link}
            to="/devices/add"
            icon={<PlusIcon className="w-4 h-4" strokeWidth={2} />}
          >
            Add Device
          </Button>
        </RestrictedAction>
      </PageHeader>

      {/* Filter bar */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 mb-5 animate-fade-in">
        <div className="flex items-center h-8 bg-card border border-border rounded-md p-0.5">
          {statusTabs.map((tab) => (
            <button
              type="button"
              key={tab.value}
              onClick={() => handleStatusChange(tab.value)}
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

        <div className="flex items-center gap-2">
          <TagFilterDropdown
            filterTags={params.tags}
            onAdd={addFilterTag}
            onRemove={removeFilterTag}
            onClearAll={clearFilterTags}
            onManageTags={() => setManageTagsOpen(true)}
          />

          <SearchField
            value={params.search}
            onChange={(next) => setSearch(next)}
            placeholder="Search by hostname..."
            aria-label="Search devices by hostname"
          />
        </div>
      </div>

      {/* Active tag filters */}
      {params.tags.length > 0 && (
        <div className="flex items-center gap-2 mb-4 animate-fade-in">
          <span className="text-2xs font-mono text-text-muted uppercase tracking-wider shrink-0">
            Filtering by:
          </span>
          <div className="flex items-center gap-1.5 flex-wrap">
            {params.tags.map((tag) => (
              <span
                key={tag}
                className="inline-flex items-center gap-1 px-2 py-0.5 bg-primary/15 text-primary text-2xs rounded-md font-medium border border-primary/20"
              >
                <TagIcon className="w-2.5 h-2.5" strokeWidth={2} />
                {tag}
                <IconButton
                  size="sm"
                  aria-label="Remove tag filter"
                  className="ml-0.5"
                  onClick={() => removeFilterTag(tag)}
                >
                  <XMarkIcon className="w-2.5 h-2.5" strokeWidth={2.5} />
                </IconButton>
              </span>
            ))}
            <Button variant="ghost" size="sm" onClick={clearFilterTags}>
              Clear all
            </Button>
          </div>
        </div>
      )}

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
        onRowClick={(device) => void navigate(`/devices/${device.uid}`)}
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

      <DeviceActionsPortal controller={deviceActions} />

      <ConnectDrawer
        open={!!connectTarget}
        onClose={() => setConnectTarget(null)}
        deviceUid={connectTarget?.uid ?? ""}
        deviceName={connectTarget?.name ?? ""}
        sshid={connectTarget?.sshid ?? ""}
      />

      <ManageTagsDrawer
        open={manageTagsOpen}
        onClose={() => {
          setManageTagsOpen(false);
          void refetch();
        }}
        onTagRenamed={(oldName, newName) => {
          mapArrayFilter("tags", (tags) =>
            tags.map((t) => (t === oldName ? newName : t)),
          );
        }}
        onTagDeleted={(name) => {
          mapArrayFilter("tags", (tags) =>
            tags.filter((t) => t !== name),
          );
        }}
      />
    </div>
  );
}
