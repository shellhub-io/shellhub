import { useState, useEffect, useMemo } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useContainers, type NormalizedContainer } from "@/hooks/useContainers";
import type { DeviceStatus } from "@/client";
import { useNamespace } from "@/hooks/useNamespaces";
import { useAuthStore } from "@/stores/authStore";
import { useTerminalStore } from "@/stores/terminalStore";
import PageHeader from "@/components/common/PageHeader";
import ConnectDrawer from "@/components/ConnectDrawer";
import ManageTagsDrawer from "@/components/ManageTagsDrawer";
import CopyButton from "@/components/common/CopyButton";
import DataTable, { type Column } from "@/components/common/DataTable";
import TagFilterDropdown from "@/components/common/TagFilterDropdown";
import { formatRelative } from "@/utils/date";
import { buildSshid } from "@/utils/sshid";
import ContainerTagsPopover from "./ContainerTagsPopover";
import ContainerActionDialog from "./ContainerActionDialog";
import AddDockerConnectorDrawer from "./AddDockerConnectorDrawer";
import {
  PlusIcon,
  MagnifyingGlassIcon,
  TagIcon,
  XMarkIcon,
  ExclamationCircleIcon,
  CubeIcon,
  ChevronDoubleRightIcon,
} from "@heroicons/react/24/outline";
import RestrictedAction from "@/components/common/RestrictedAction";

const statusTabs: { label: string; value: DeviceStatus }[] = [
  { label: "Accepted", value: "accepted" },
  { label: "Pending", value: "pending" },
  { label: "Rejected", value: "rejected" },
];

const PER_PAGE = 10;
const SEARCH_DEBOUNCE_MS = 300;
const VALID_STATUSES = new Set<string>(["accepted", "pending", "rejected"]);

export default function Containers() {
  const [searchParams] = useSearchParams();
  const initialStatus = searchParams.get("status") ?? "accepted";
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState<DeviceStatus>(
    VALID_STATUSES.has(initialStatus)
      ? (initialStatus as DeviceStatus)
      : "accepted",
  );
  const [filterTags, setFilterTags] = useState<string[]>([]);
  const [searchInput, setSearchInput] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [actionTarget, setActionTarget] = useState<{
    container: NormalizedContainer;
    action: "accept" | "reject" | "remove";
  } | null>(null);
  const [connectTarget, setConnectTarget] = useState<{
    uid: string;
    name: string;
    sshid: string;
  } | null>(null);
  const [manageTagsOpen, setManageTagsOpen] = useState(false);
  const [addConnectorOpen, setAddConnectorOpen] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchInput.trim());
    }, SEARCH_DEBOUNCE_MS);
    return () => clearTimeout(timer);
  }, [searchInput]);

  const { containers, totalCount, isLoading, error, refetch } = useContainers({
    page,
    perPage: PER_PAGE,
    status,
    search: debouncedSearch,
    filterTags,
  });

  const tenantId = useAuthStore((s) => s.tenant) ?? "";
  const { namespace: currentNamespace } = useNamespace(tenantId);
  const navigate = useNavigate();

  const totalPages = Math.ceil(totalCount / PER_PAGE);
  const nsName = currentNamespace?.name ?? "";

  const handleStatusChange = (newStatus: DeviceStatus) => {
    setStatus(newStatus);
    setPage(1);
  };

  const addFilterTag = (tag: string) => {
    setFilterTags((prev) => (prev.includes(tag) ? prev : [...prev, tag]));
    setPage(1);
  };

  const removeFilterTag = (tag: string) => {
    setFilterTags((prev) => prev.filter((t) => t !== tag));
    setPage(1);
  };

  const clearFilterTags = () => {
    setFilterTags([]);
    setPage(1);
  };

  const columns = useMemo<Column<NormalizedContainer>[]>(() => {
    const baseColumns: Column<NormalizedContainer>[] = [
      {
        key: "hostname",
        header: "Hostname",
        render: (container) => (
          <span className="text-sm font-medium text-text-primary group-hover:text-primary transition-colors">
            {container.name}
          </span>
        ),
      },
      {
        key: "image",
        header: "Image",
        render: (container) => (
          <span className="text-xs text-text-secondary font-mono truncate max-w-[200px] block">
            {container.info?.pretty_name ?? "Unknown"}
          </span>
        ),
      },
      {
        key: "tags",
        header: "Tags",
        render: (container) => (
          <ContainerTagsPopover
            container={container}
            onFilterTag={addFilterTag}
          />
        ),
      },
      {
        key: "last_seen",
        header: "Last Seen",
        render: (container) => (
          <span className="text-xs text-text-secondary">
            {formatRelative(container.last_seen)}
          </span>
        ),
      },
    ];

    if (status === "accepted") {
      return [
        {
          key: "online",
          header: "",
          headerClassName: "w-12",
          render: (container) =>
            container.online ? (
              <span className="relative flex h-2.5 w-2.5 mx-auto">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent-green opacity-40" />
                <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-accent-green shadow-[0_0_6px_rgba(130,165,104,0.4)]" />
              </span>
            ) : (
              <span className="block w-2.5 h-2.5 rounded-full mx-auto bg-text-muted/30" />
            ),
        },
        baseColumns[0], // hostname
        {
          key: "sshid",
          header: "SSHID",
          render: (container) => {
            const sshid = nsName
              ? buildSshid(nsName, container.name)
              : container.uid.substring(0, 8);
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
        ...baseColumns.slice(1), // image, tags, last_seen
        {
          key: "connect",
          header: "",
          headerClassName: "w-20",
          render: (container) =>
            container.online ? (
              <RestrictedAction action="device:connect">
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    const existing = useTerminalStore
                      .getState()
                      .sessions.find((s) => s.deviceUid === container.uid);
                    if (existing) {
                      useTerminalStore.getState().restore(existing.id);
                    } else {
                      const sshid = nsName
                        ? buildSshid(nsName, container.name)
                        : container.uid;
                      setConnectTarget({
                        uid: container.uid,
                        name: container.name,
                        sshid,
                      });
                    }
                  }}
                  className="inline-flex items-center gap-1 px-2.5 py-1 bg-accent-green/10 text-accent-green text-2xs font-semibold rounded-md hover:bg-accent-green/20 border border-accent-green/20 transition-all"
                >
                  <ChevronDoubleRightIcon className="w-3 h-3" strokeWidth={2} />
                  Connect
                </button>
              </RestrictedAction>
            ) : (
              <span className="text-2xs text-text-muted/30 font-mono">
                Offline
              </span>
            ),
        },
      ];
    }

    if (status === "pending") {
      return [
        ...baseColumns,
        {
          key: "actions",
          header: "Actions",
          headerClassName: "text-right",
          render: (container) => (
            <div className="flex items-center justify-end gap-1.5">
              <RestrictedAction action="device:accept">
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    setActionTarget({ container, action: "accept" });
                  }}
                  className="px-2.5 py-1 text-2xs font-semibold rounded-md bg-accent-green/10 text-accent-green hover:bg-accent-green/20 border border-accent-green/20 transition-all"
                >
                  Accept
                </button>
              </RestrictedAction>
              <RestrictedAction action="device:reject">
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    setActionTarget({ container, action: "reject" });
                  }}
                  className="px-2.5 py-1 text-2xs font-semibold rounded-md bg-accent-yellow/10 text-accent-yellow hover:bg-accent-yellow/20 border border-accent-yellow/20 transition-all"
                >
                  Reject
                </button>
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
        render: (container) => (
          <div className="flex items-center justify-end gap-1.5">
            <RestrictedAction action="device:accept">
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  setActionTarget({ container, action: "accept" });
                }}
                className="px-2.5 py-1 text-2xs font-semibold rounded-md bg-accent-green/10 text-accent-green hover:bg-accent-green/20 border border-accent-green/20 transition-all"
              >
                Accept
              </button>
            </RestrictedAction>
            <RestrictedAction action="device:remove">
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  setActionTarget({ container, action: "remove" });
                }}
                className="px-2.5 py-1 text-2xs font-semibold rounded-md bg-accent-red/10 text-accent-red hover:bg-accent-red/20 border border-accent-red/20 transition-all"
              >
                Remove
              </button>
            </RestrictedAction>
          </div>
        ),
      },
    ];
  }, [status, nsName]);

  return (
    <div>
      <PageHeader
        icon={<CubeIcon className="w-6 h-6" />}
        overline="Container Management"
        title="Containers"
        description="Manage and monitor Docker containers connected via ShellHub Connector"
      >
        <button
          onClick={() => setAddConnectorOpen(true)}
          className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all duration-200"
        >
          <PlusIcon className="w-4 h-4" strokeWidth={2} />
          Add Docker Host
        </button>
      </PageHeader>

      {/* Filter bar */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 mb-5 animate-fade-in">
        <div
          className="flex items-center h-8 bg-card border border-border rounded-md p-0.5"
          role="tablist"
          aria-label="Container status filter"
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

        <div className="flex items-center gap-2">
          <TagFilterDropdown
            filterTags={filterTags}
            onAdd={addFilterTag}
            onRemove={removeFilterTag}
            onClearAll={clearFilterTags}
            onManageTags={() => setManageTagsOpen(true)}
          />

          <div className="relative h-8">
            <MagnifyingGlassIcon
              className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-text-muted"
              strokeWidth={2}
            />
            <input
              type="text"
              value={searchInput}
              onChange={(e) => {
                setSearchInput(e.target.value);
                setPage(1);
              }}
              placeholder="Search containers..."
              aria-label="Search containers"
              className="h-full pl-9 pr-3 bg-card border border-border rounded-md text-xs text-text-primary font-mono placeholder:text-text-secondary focus:outline-none focus:border-primary/40 focus:ring-1 focus:ring-primary/15 transition-all duration-200 w-56"
            />
          </div>
        </div>
      </div>

      {/* Active tag filters */}
      {filterTags.length > 0 && (
        <div className="flex items-center gap-2 mb-4 animate-fade-in">
          <span className="text-2xs font-mono text-text-muted uppercase tracking-wider shrink-0">
            Filtering by:
          </span>
          <div className="flex items-center gap-1.5 flex-wrap">
            {filterTags.map((tag) => (
              <span
                key={tag}
                className="inline-flex items-center gap-1 px-2 py-0.5 bg-primary/15 text-primary text-2xs rounded-md font-medium border border-primary/20"
              >
                <TagIcon className="w-2.5 h-2.5" strokeWidth={2} />
                {tag}
                <button
                  onClick={() => removeFilterTag(tag)}
                  aria-label={`Remove ${tag} filter`}
                  className="hover:text-white transition-colors ml-0.5"
                >
                  <XMarkIcon className="w-2.5 h-2.5" strokeWidth={2.5} />
                </button>
              </span>
            ))}
            <button
              onClick={clearFilterTags}
              className="text-2xs text-text-muted hover:text-text-primary transition-colors font-mono"
            >
              Clear all
            </button>
          </div>
        </div>
      )}

      {error && (
        <div className="flex items-center gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3.5 py-2.5 rounded-md text-xs font-mono mb-4 animate-slide-down">
          <ExclamationCircleIcon
            className="w-3.5 h-3.5 shrink-0"
            strokeWidth={2}
          />
          {error.message}
        </div>
      )}

      <DataTable
        columns={columns}
        data={containers}
        rowKey={(container) => container.uid}
        isLoading={isLoading}
        loadingMessage="Loading containers..."
        page={page}
        totalPages={totalPages}
        totalCount={totalCount}
        itemLabel="container"
        onPageChange={setPage}
        onRowClick={(container) =>
          void navigate(`/containers/${container.uid}`)
        }
        emptyState={
          <div className="text-center">
            <CubeIcon
              className="w-10 h-10 text-text-muted/30 mx-auto mb-3"
              strokeWidth={1}
            />
            <p className="text-xs font-mono text-text-muted">
              {debouncedSearch
                ? `No containers matching "${debouncedSearch}"`
                : "No containers found"}
            </p>
          </div>
        }
      />

      <ContainerActionDialog
        key={
          actionTarget
            ? `${actionTarget.action}/${actionTarget.container.uid}`
            : "closed"
        }
        open={!!actionTarget}
        container={actionTarget?.container ?? null}
        action={actionTarget?.action ?? "accept"}
        onClose={() => setActionTarget(null)}
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
          setFilterTags((prev) =>
            prev.map((t) => (t === oldName ? newName : t)),
          );
        }}
        onTagDeleted={(name) => {
          setFilterTags((prev) => prev.filter((t) => t !== name));
        }}
      />

      <AddDockerConnectorDrawer
        open={addConnectorOpen}
        onClose={() => setAddConnectorOpen(false)}
      />
    </div>
  );
}
