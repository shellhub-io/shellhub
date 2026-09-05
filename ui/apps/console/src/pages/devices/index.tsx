import { useState, useMemo, useCallback } from "react";
import { useNavigate, Link } from "react-router-dom";
import { useDevices, type NormalizedDevice } from "@/hooks/useDevices";
import { useDebouncedValue } from "@/hooks/useDebouncedValue";
import { useTableSort } from "@/hooks/useTableSort";
import { usePaginatedListState } from "@/hooks/usePaginatedListState";
import { useNamespace } from "@/hooks/useNamespaces";
import { useAuthStore } from "@/stores/authStore";
import { useTerminalStore } from "@/stores/terminalStore";
import PageHeader from "@/components/common/PageHeader";
import ConnectDrawer from "@/components/ConnectDrawer";
import ManageTagsDrawer from "@/components/ManageTagsDrawer";
import CopyButton from "@/components/common/CopyButton";
import PlatformBadge from "@/components/common/PlatformBadge";
import OnlineDot from "@/components/common/OnlineDot";
import LastSeenCell from "@/components/common/LastSeenCell";
import DataTable, { type Column } from "@/components/common/DataTable";
import SearchField from "@/components/common/fields/SearchField";
import { buildSshid } from "@/utils/sshid";
import TagFilterDropdown from "@/components/common/TagFilterDropdown";
import TagsPopover from "@/components/common/TagsPopover";
import {
  useAddDeviceTag,
  useRemoveDeviceTag,
} from "@/hooks/useDeviceMutations";
import {
  PlusIcon,
  TagIcon,
  ArrowRightIcon,
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
import { apiErrorMessage } from "@/api/errors";
import { PER_PAGE, pageCount } from "@/utils/pagination";

const SEARCH_DEBOUNCE_MS = 300;

const relatedPages: { label: string; to: string }[] = [
  { label: "Pending", to: "/pending-devices" },
  { label: "Install Keys", to: "/install-keys" },
];

type DevicesParams = {
  page: number;
  search: string;
  tags: string[];
};

const DEFAULTS: DevicesParams = {
  page: 1,
  search: "",
  tags: [],
};

type SortField = "name" | "last_seen";

/**
 * The devices list — the console's main page — filtered by status, search and tags, with the
 * filter state held in the URL so a view can be shared.
 */
export default function Devices() {
  const { params, setPage, setSearch, setArrayFilter, mapArrayFilter } =
    usePaginatedListState<DevicesParams>({
      defaults: DEFAULTS,
    });

  const debouncedSearch = useDebouncedValue(
    params.search.trim(),
    SEARCH_DEBOUNCE_MS,
  );

  const addDeviceTag = useAddDeviceTag();
  const removeDeviceTag = useRemoveDeviceTag();
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
    status: "accepted",
    search: debouncedSearch,
    filterTags: params.tags,
    sortBy,
    orderBy,
  });

  const tenantId = useAuthStore((s) => s.tenant) ?? "";
  const { namespace: currentNamespace } = useNamespace(tenantId);
  const navigate = useNavigate();

  const totalPages = pageCount(totalCount);
  const nsName = currentNamespace?.name ?? "";

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
    const [hostnameColumn, ...detailColumns]: Column<NormalizedDevice>[] = [
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
          <TagsPopover
            uid={device.uid}
            tags={device.tags}
            addTag={addDeviceTag.mutateAsync}
            removeTag={removeDeviceTag.mutateAsync}
            onFilterTag={addFilterTag}
          />
        ),
      },
      {
        key: "last_seen",
        header: "Last Seen",
        sortable: true,
        render: (device) => <LastSeenCell value={device.last_seen} />,
      },
    ];

    return [
      {
        key: "online",
        header: "",
        headerClassName: "w-12",
        render: (device) => <OnlineDot online={device.online} />,
      },
      hostnameColumn,
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
      ...detailColumns,
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
                  <ChevronDoubleRightIcon className="w-3 h-3" strokeWidth={2} />
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
  }, [
    nsName,
    addFilterTag,
    addDeviceTag.mutateAsync,
    removeDeviceTag.mutateAsync,
  ]);

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
        <div className="flex items-center gap-3">
          <div className="flex items-center h-8 bg-card border border-border rounded-md p-0.5">
            <span className="h-full inline-flex items-center px-3.5 text-xs font-medium rounded bg-primary/15 text-primary border border-primary/25">
              Accepted
            </span>
            {relatedPages.map((link) => (
              <Link
                key={link.to}
                to={link.to}
                className="h-full inline-flex items-center gap-1 px-3.5 text-xs font-medium rounded border border-transparent text-text-muted transition-all duration-150 hover:text-primary hover:bg-primary/10"
              >
                {link.label}
                <ArrowRightIcon className="w-3 h-3" strokeWidth={2.5} />
              </Link>
            ))}
          </div>
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
          {apiErrorMessage(error)}
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
          mapArrayFilter("tags", (tags) => tags.filter((t) => t !== name));
        }}
      />
    </div>
  );
}
