import { useEffect, useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { useDevicesStore } from "../../stores/devicesStore";
import { useNamespacesStore } from "../../stores/namespacesStore";
import { useTerminalStore } from "../../stores/terminalStore";
import { Device } from "../../types/device";
import PageHeader from "../../components/common/PageHeader";
import ConnectDrawer from "../../components/ConnectDrawer";
import ManageTagsDrawer from "../../components/ManageTagsDrawer";
import CopyButton from "../../components/common/CopyButton";
import PlatformBadge from "../../components/common/PlatformBadge";
import { formatRelative } from "../../utils/date";
import { TH as TH_BASE } from "../../utils/styles";
import Pagination from "../../components/common/Pagination";
import TagFilterDropdown from "./TagFilterDropdown";
import TagsPopover from "./TagsPopover";
import DeviceActionDialog from "./DeviceActionDialog";
import {
  PlusIcon,
  MagnifyingGlassIcon,
  TagIcon,
  XMarkIcon,
  ExclamationCircleIcon,
  CpuChipIcon,
  ChevronDoubleRightIcon,
} from "@heroicons/react/24/outline";

const statusTabs = [
  { label: "Accepted", value: "accepted" },
  { label: "Pending", value: "pending" },
  { label: "Rejected", value: "rejected" },
];

const TH = `${TH_BASE} whitespace-nowrap`;

/* ─── Page ─── */
export default function Devices() {
  const {
    devices,
    totalCount,
    loading,
    error,
    page,
    perPage,
    status,
    filterTags,
    fetch,
    setPage,
    setStatus,
    addFilterTag,
    removeFilterTag,
    clearFilterTags,
  } = useDevicesStore();
  const { currentNamespace } = useNamespacesStore();
  const navigate = useNavigate();
  const [search, setSearch] = useState("");
  const [actionTarget, setActionTarget] = useState<{
    device: Device;
    action: "accept" | "reject" | "remove";
  } | null>(null);
  const [connectTarget, setConnectTarget] = useState<{
    uid: string;
    name: string;
    sshid: string;
  } | null>(null);
  const [manageTagsOpen, setManageTagsOpen] = useState(false);

  useEffect(() => {
    fetch();
  }, [fetch]);

  const totalPages = Math.ceil(totalCount / perPage);
  const nsName = currentNamespace?.name ?? "";

  const handleStatusChange = (newStatus: string) => {
    setStatus(newStatus);
    fetch(1, perPage, newStatus);
  };

  const filtered = search
    ? devices.filter(
        (d) =>
          d.name.toLowerCase().includes(search.toLowerCase()) ||
          d.uid.toLowerCase().includes(search.toLowerCase()),
      )
    : devices;

  return (
    <div>
      <PageHeader
        icon={<CpuChipIcon className="w-6 h-6" />}
        overline="Device Management"
        title="Devices"
        description="Manage and monitor all devices connected to your namespace"
      >
        <Link
          to="/devices/add"
          className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all duration-200"
        >
          <PlusIcon className="w-4 h-4" strokeWidth={2} />
          Add Device
        </Link>
      </PageHeader>

      {/* Filter bar */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 mb-5 animate-fade-in">
        <div className="flex items-center h-8 bg-card border border-border rounded-md p-0.5">
          {statusTabs.map((tab) => (
            <button
              key={tab.value}
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
            onManageTags={() => setManageTagsOpen(true)}
          />

          <div className="relative h-8">
            <MagnifyingGlassIcon
              className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-text-muted"
              strokeWidth={2}
            />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search devices..."
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
          {error}
        </div>
      )}

      {/* Table */}
      <div className="bg-card border border-border rounded-xl overflow-hidden animate-fade-in">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-border bg-surface/50">
                {status === "accepted" && <th className={`${TH} w-12`} />}
                <th className={TH}>Hostname</th>
                {status === "accepted" && <th className={TH}>SSH ID</th>}
                <th className={TH}>Operating System</th>
                <th className={TH}>Tags</th>
                <th className={TH}>Last Seen</th>
                {status === "accepted" && <th className={`${TH} w-20`} />}
                {status !== "accepted" && (
                  <th className={`${TH} text-right`}>Actions</th>
                )}
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {loading && devices.length === 0 ? (
                <tr>
                  <td
                    colSpan={status === "accepted" ? 7 : 5}
                    className="px-4 py-16 text-center"
                  >
                    <div className="flex items-center justify-center gap-3">
                      <span className="w-4 h-4 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
                      <span className="text-xs font-mono text-text-muted">
                        Loading devices...
                      </span>
                    </div>
                  </td>
                </tr>
              ) : filtered.length === 0 ? (
                <tr>
                  <td
                    colSpan={status === "accepted" ? 7 : 5}
                    className="px-4 py-16 text-center"
                  >
                    <p className="text-xs font-mono text-text-muted">
                      {search
                        ? `No devices matching "${search}"`
                        : "No devices found"}
                    </p>
                  </td>
                </tr>
              ) : (
                filtered.map((device) => {
                  const sshid = nsName
                    ? `${nsName}.${device.name}@${nsName}`
                    : device.uid.substring(0, 8);
                  return (
                    <tr
                      key={device.uid}
                      onClick={() => {
                        if (device.status === "accepted")
                          navigate(`/devices/${device.uid}`);
                      }}
                      className={`group hover:bg-hover-subtle transition-colors ${device.status === "accepted" ? "cursor-pointer" : ""}`}
                    >
                      {/* Online dot — accepted only */}
                      {status === "accepted" && (
                        <td className="px-4 py-3.5 w-12">
                          {device.online ? (
                            <span className="relative flex h-2.5 w-2.5 mx-auto">
                              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent-green opacity-40" />
                              <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-accent-green shadow-[0_0_6px_rgba(130,165,104,0.4)]" />
                            </span>
                          ) : (
                            <span className="block w-2.5 h-2.5 rounded-full mx-auto bg-text-muted/30" />
                          )}
                        </td>
                      )}

                      {/* Hostname */}
                      <td className="px-4 py-3.5">
                        <span className="text-sm font-medium text-text-primary group-hover:text-primary transition-colors">
                          {device.name}
                        </span>
                      </td>

                      {/* SSH ID — accepted only */}
                      {status === "accepted" && (
                        <td className="px-4 py-3.5">
                          <div className="flex items-center gap-1">
                            <code
                              className="text-2xs font-mono text-text-muted truncate max-w-[220px]"
                              title={sshid}
                            >
                              {sshid}
                            </code>
                            <CopyButton text={sshid} />
                          </div>
                        </td>
                      )}

                      {/* OS + Platform */}
                      <td className="px-4 py-3.5">
                        <div className="flex items-center gap-2">
                          <span className="text-xs text-text-secondary truncate max-w-[160px]">
                            {device.info?.pretty_name ?? "Unknown"}
                          </span>
                          {device.info?.platform && (
                            <PlatformBadge platform={device.info.platform} />
                          )}
                        </div>
                      </td>

                      {/* Tags */}
                      <td className="px-4 py-3.5">
                        <TagsPopover
                          device={device}
                          onUpdated={() => fetch()}
                          onFilterTag={addFilterTag}
                        />
                      </td>

                      {/* Last Seen */}
                      <td className="px-4 py-3.5">
                        <span className="text-xs text-text-secondary">
                          {formatRelative(device.last_seen)}
                        </span>
                      </td>

                      {/* Connect — accepted only */}
                      {status === "accepted" && (
                        <td className="px-4 py-3.5 w-20">
                          {device.online ? (
                            <button
                              onClick={(e) => {
                                e.stopPropagation();
                                const existing = useTerminalStore
                                  .getState()
                                  .sessions.find(
                                    (s) => s.deviceUid === device.uid,
                                  );
                                if (existing) {
                                  useTerminalStore
                                    .getState()
                                    .restore(existing.id);
                                } else {
                                  const sshid = nsName
                                    ? `${nsName}.${device.name}@${nsName}`
                                    : device.uid;
                                  setConnectTarget({
                                    uid: device.uid,
                                    name: device.name,
                                    sshid,
                                  });
                                }
                              }}
                              className="inline-flex items-center gap-1 px-2.5 py-1 bg-accent-green/10 text-accent-green text-2xs font-semibold rounded-md hover:bg-accent-green/20 border border-accent-green/20 transition-all"
                            >
                              <ChevronDoubleRightIcon
                                className="w-3 h-3"
                                strokeWidth={2}
                              />
                              Connect
                            </button>
                          ) : (
                            <span className="text-2xs text-text-muted/30 font-mono">
                              Offline
                            </span>
                          )}
                        </td>
                      )}

                      {/* Actions — pending/rejected only */}
                      {status === "pending" && (
                        <td className="px-4 py-3.5 text-right">
                          <div className="flex items-center justify-end gap-1.5">
                            <button
                              onClick={(e) => {
                                e.stopPropagation();
                                setActionTarget({ device, action: "accept" });
                              }}
                              className="px-2.5 py-1 text-2xs font-semibold rounded-md bg-accent-green/10 text-accent-green hover:bg-accent-green/20 border border-accent-green/20 transition-all"
                            >
                              Accept
                            </button>
                            <button
                              onClick={(e) => {
                                e.stopPropagation();
                                setActionTarget({ device, action: "reject" });
                              }}
                              className="px-2.5 py-1 text-2xs font-semibold rounded-md bg-accent-yellow/10 text-accent-yellow hover:bg-accent-yellow/20 border border-accent-yellow/20 transition-all"
                            >
                              Reject
                            </button>
                          </div>
                        </td>
                      )}
                      {status === "rejected" && (
                        <td className="px-4 py-3.5 text-right">
                          <div className="flex items-center justify-end gap-1.5">
                            <button
                              onClick={(e) => {
                                e.stopPropagation();
                                setActionTarget({ device, action: "accept" });
                              }}
                              className="px-2.5 py-1 text-2xs font-semibold rounded-md bg-accent-green/10 text-accent-green hover:bg-accent-green/20 border border-accent-green/20 transition-all"
                            >
                              Accept
                            </button>
                            <button
                              onClick={(e) => {
                                e.stopPropagation();
                                setActionTarget({ device, action: "remove" });
                              }}
                              className="px-2.5 py-1 text-2xs font-semibold rounded-md bg-accent-red/10 text-accent-red hover:bg-accent-red/20 border border-accent-red/20 transition-all"
                            >
                              Remove
                            </button>
                          </div>
                        </td>
                      )}
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>

      <Pagination
        page={page}
        totalPages={totalPages}
        totalCount={totalCount}
        itemLabel="device"
        onPageChange={(p) => {
          setPage(p);
          fetch(p);
        }}
      />

      {/* Action Dialog */}
      {actionTarget && (
        <DeviceActionDialog
          device={actionTarget.device}
          action={actionTarget.action}
          onClose={() => setActionTarget(null)}
        />
      )}

      {/* Connect Drawer */}
      <ConnectDrawer
        open={!!connectTarget}
        onClose={() => setConnectTarget(null)}
        deviceUid={connectTarget?.uid ?? ""}
        deviceName={connectTarget?.name ?? ""}
        sshid={connectTarget?.sshid ?? ""}
      />

      {/* Manage Tags Drawer */}
      <ManageTagsDrawer
        open={manageTagsOpen}
        onClose={() => setManageTagsOpen(false)}
      />
    </div>
  );
}
