import { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { devicesQueryOptions, type NormalizedDevice } from "@/hooks/useDevices";
import { useTableSort } from "@/hooks/useTableSort";
import { usePagedList } from "@/hooks/usePagedList";
import { useNamespace } from "@/hooks/useNamespaceRequests";
import { useFirstRunEligible } from "@/hooks/useFirstRunEligible";
import { useAuthStore } from "@/stores/authStore";
import { useTerminalStore } from "@/stores/terminalStore";
import PageHeader from "@/components/common/PageHeader";
import ManageTagsModal from "@/components/ManageTagsModal";
import CopyButton from "@/components/common/CopyButton";
import PlatformBadge from "@/components/common/PlatformBadge";
import OnlineDot from "@/components/common/OnlineDot";
import LastSeenCell from "@/components/common/LastSeenCell";
import DataTable, { type Column } from "@/components/common/DataTable";
import SearchField from "@/components/common/fields/SearchField";
import { buildSshid } from "@/utils/sshid";
import TagsPopover from "@/components/common/TagsPopover";
import {
  useAddDeviceTag,
  useRemoveDeviceTag,
  useRenameDevice,
} from "@/hooks/useDeviceRequests";
import { useHasPermission } from "@/hooks/useHasPermission";
import {
  PlusIcon,
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
import RenameableName from "@/components/common/RenameableName";
import { apiErrorMessage } from "@/api/errors";
import { PER_PAGE } from "@/utils/pagination";
import { useNavSectionTitle } from "@/components/layout/navSections";
import WaitingForApproval from "./WaitingForApproval";
import DeviceFilters from "./DeviceFilters";

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
  const sectionTitle = useNavSectionTitle("/devices");
  const { sortBy, orderBy, handleSort } = useTableSort<SortField>({
    defaultField: "last_seen",
    onSortChange: () => setPage(1),
  });
  const {
    params,
    appliedSearch: searchTerm,
    setPage,
    setSearch,
    mapArrayFilter,
    rows: devices,
    isLoading,
    error,
    refetch,
    setFilter,
    pagination,
  } = usePagedList({
    defaults: DEFAULTS,
    query: ({ page, search, tags }: DevicesParams) =>
      devicesQueryOptions({
        page,
        perPage: PER_PAGE,
        status: "accepted",
        search,
        filterTags: tags,
        sortBy,
        orderBy,
      }),
  });

  const firstRunOpen =
    useFirstRunEligible() === true && !searchTerm && params.tags.length === 0;

  const addDeviceTag = useAddDeviceTag();
  const removeDeviceTag = useRemoveDeviceTag();
  const renameDevice = useRenameDevice();
  const canRename = useHasPermission("device:rename");
  const [manageTagsOpen, setManageTagsOpen] = useState(false);

  const tenantId = useAuthStore((s) => s.tenant) ?? "";
  const { namespace: currentNamespace } = useNamespace(tenantId);
  const navigate = useNavigate();

  const nsName = currentNamespace?.name ?? "";

  const addFilterTag = (tag: string) => {
    mapArrayFilter("tags", (tags) =>
      tags.includes(tag) ? tags : [...tags, tag],
    );
  };

  const removeFilterTag = (tag: string) => {
    mapArrayFilter("tags", (tags) => tags.filter((t) => t !== tag));
  };

  const clearFilterTags = () => {
    setFilter("tags", []);
  };

  const columns = ((): Column<NormalizedDevice>[] => {
    const [hostnameColumn, ...detailColumns]: Column<NormalizedDevice>[] = [
      {
        key: "name",
        header: "Hostname",
        sortable: true,
        render: (device) => (
          <RenameableName
            uid={device.uid}
            name={device.name}
            entityLabel="device"
            rename={renameDevice.mutateAsync}
            canRename={canRename}
            nsName={nsName}
          />
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
                  useTerminalStore.getState().connect(device.uid, device.name);
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
  })();

  return (
    <div>
      <PageHeader
        icon={<CpuChipIcon className="w-6 h-6" />}
        overline={sectionTitle}
        title="Devices"
        description="Manage and monitor all devices connected to your namespace"
      />

      <div className="flex flex-col sm:flex-row sm:items-center gap-2 mb-5 animate-fade-in">
        <SearchField
          full
          toolbar
          className="flex-1 min-w-0"
          value={params.search}
          onChange={(next) => setSearch(next)}
          placeholder="Search by hostname..."
          aria-label="Search devices by hostname"
        />

        <div className="flex items-center gap-2">
          <DeviceFilters
            filterTags={params.tags}
            onAdd={addFilterTag}
            onRemove={removeFilterTag}
            onClearAll={clearFilterTags}
            onManageTags={() => setManageTagsOpen(true)}
          />

          <WaitingForApproval />

          <RestrictedAction action="device:add">
            <Button
              as={Link}
              to="/devices/add"
              size="sm"
              className="h-9"
              icon={<PlusIcon className="w-4 h-4" strokeWidth={2} />}
            >
              Add Device
            </Button>
          </RestrictedAction>
        </div>
      </div>

      {params.tags.length > 0 && (
        <div className="flex items-center gap-1.5 flex-wrap -mt-2 mb-4 animate-fade-in">
          {params.tags.map((tag) => (
            <span
              key={tag}
              className="inline-flex items-center gap-1.5 h-7 pl-2.5 pr-1 text-xs rounded-md border border-primary/25 bg-primary/10 text-primary"
            >
              <span className="text-text-secondary">Tag:</span>
              {tag}
              <IconButton
                size="sm"
                aria-label={`Remove the ${tag} filter`}
                onClick={() => removeFilterTag(tag)}
              >
                <XMarkIcon className="w-3 h-3" strokeWidth={2.5} />
              </IconButton>
            </span>
          ))}
          <Button variant="ghost" size="sm" onClick={clearFilterTags}>
            Clear all
          </Button>
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
        {...pagination}
        itemLabel="device"
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
              {searchTerm
                ? `No devices matching "${searchTerm}"`
                : "No devices found"}
            </p>
            {firstRunOpen && (
              <Link
                to="/dashboard"
                className="inline-block mt-3 text-xs font-medium text-primary hover:text-primary-400 transition-colors"
              >
                Continue setup
              </Link>
            )}
          </div>
        }
      />

      <ManageTagsModal
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
