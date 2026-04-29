import { useEffect, useState } from "react";
import {
  useParams,
  useNavigate,
  useSearchParams,
  Link,
} from "react-router-dom";
import {
  TagIcon,
  XMarkIcon,
  PlusIcon,
  PencilSquareIcon,
  CheckIcon,
  ChevronRightIcon,
  TrashIcon,
  InformationCircleIcon,
  ComputerDesktopIcon,
  ClockIcon,
  CpuChipIcon,
  ChevronDoubleRightIcon,
} from "@heroicons/react/24/outline";
import { useDevice } from "../hooks/useDevice";
import {
  useRenameDevice,
  useAddDeviceTag,
  useRemoveDeviceTag,
  useRemoveDevice,
  useUpdateDeviceCustomFields,
} from "../hooks/useDeviceMutations";
import { useNamespace } from "../hooks/useNamespaces";
import { useAuthStore } from "../stores/authStore";
import { useTerminalStore } from "../stores/terminalStore";
import DeviceActionDialog from "./devices/DeviceActionDialog";
import ConnectDrawer from "../components/ConnectDrawer";
import CopyButton from "../components/common/CopyButton";
import PlatformBadge from "../components/common/PlatformBadge";
import { formatDateFull, formatRelative } from "../utils/date";
import { buildSshid } from "../utils/sshid";
import { useHasPermission } from "../hooks/useHasPermission";
import RestrictedAction from "../components/common/RestrictedAction";

/* ─── Shared styles ─── */
const LABEL
  = "text-2xs font-mono font-semibold uppercase tracking-label text-text-muted";
const VALUE = "text-sm text-text-primary font-medium mt-0.5";

/* ─── Info Row ─── */
function InfoItem({
  label,
  value,
  mono,
  copyable,
  truncate,
}: {
  label: string;
  value: string;
  mono?: boolean;
  copyable?: boolean;
  truncate?: number;
}) {
  const display = truncate && value ? value.slice(0, truncate) : value;

  return (
    <div>
      <dt className={LABEL}>{label}</dt>
      <dd className="flex items-center gap-1 mt-0.5">
        <span
          className={`text-sm text-text-primary ${mono ? "font-mono text-xs" : "font-medium"}`}
        >
          {display || "—"}
        </span>
        {copyable && value && <CopyButton text={value} />}
      </dd>
    </div>
  );
}

/* ─── Tags Section ─── */
function TagsSection({ uid, tags }: { uid: string; tags: string[] }) {
  const addTagMutation = useAddDeviceTag();
  const removeTagMutation = useRemoveDeviceTag();
  const canEditTags = useHasPermission("tag:edit");
  const [input, setInput] = useState("");
  const [adding, setAdding] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleAdd = async () => {
    const tag = input.trim();
    if (!tag) return;
    setError(null);

    if (tags && tags.includes(tag)) {
      setError("This tag is already added.");
      return;
    }
    if (tags && tags.length >= 3) return;
    if (tag.length < 3) {
      setError("Tag must be at least 3 characters.");
      return;
    }
    if (tag.length > 255) {
      setError("Tag must be at most 255 characters.");
      return;
    }
    if (!/^[a-zA-Z0-9]+$/.test(tag)) {
      setError("Tag must contain only letters and numbers.");
      return;
    }

    setAdding(true);
    try {
      await addTagMutation.mutateAsync({ path: { uid, name: tag } });
      setInput("");
    } catch {
      setError("Failed to add tag.");
    }
    setAdding(false);
  };

  const handleRemove = async (tag: string) => {
    try {
      await removeTagMutation.mutateAsync({ path: { uid, name: tag } });
    } catch {
      /* invalidation handles UI update */
    }
  };

  return (
    <div>
      <h3 className={LABEL + " mb-2"}>Tags</h3>
      <div className="flex flex-wrap items-center gap-2">
        {tags
          && tags.map((tag) => (
            <span
              key={tag}
              className="inline-flex items-center gap-1.5 px-2.5 py-1 bg-primary/10 text-primary text-xs rounded-md font-medium"
            >
              <TagIcon className="w-3 h-3" strokeWidth={2} />
              {tag}
              {canEditTags && (
                <button
                  onClick={() => void handleRemove(tag)}
                  className="hover:text-white transition-colors"
                >
                  <XMarkIcon className="w-3 h-3" strokeWidth={2} />
                </button>
              )}
            </span>
          ))}
        {canEditTags && (!tags || tags.length < 3) && (
          <div className="flex items-center gap-1.5">
            <input
              type="text"
              value={input}
              onChange={(e) => {
                setInput(e.target.value);
                setError(null);
              }}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  void handleAdd();
                }
              }}
              placeholder="Add tag..."
              pattern="^[a-zA-Z0-9]+$"
              className="w-28 px-2.5 py-1 bg-card border border-border rounded-md text-xs text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-primary/40 transition-all"
            />
            <button
              onClick={() => void handleAdd()}
              disabled={adding || !input.trim()}
              className="p-1 rounded-md text-text-muted hover:text-primary hover:bg-primary/10 disabled:opacity-soft transition-all"
            >
              <PlusIcon className="w-4 h-4" strokeWidth={2} />
            </button>
          </div>
        )}
      </div>
      {tags && tags.length >= 3 && (
        <p className="text-2xs text-text-muted mt-1.5">
          Maximum of 3 tags reached.
        </p>
      )}
      {error && (
        <p className="text-2xs text-accent-red mt-1.5">{error}</p>
      )}
    </div>
  );
}

/* ─── Rename Inline ─── */
function RenameSection({
  uid,
  currentName,
}: {
  uid: string;
  currentName: string;
}) {
  const renameMutation = useRenameDevice();
  const canRename = useHasPermission("device:rename");
  const [editing, setEditing] = useState(false);
  const [name, setName] = useState(currentName);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSave = async () => {
    if (!name.trim() || name.trim() === currentName) {
      setEditing(false);
      return;
    }
    setSaving(true);
    setError(null);
    try {
      await renameMutation.mutateAsync({ path: { uid }, body: { name: name.trim() } });
      setEditing(false);
    } catch {
      setError("Failed to rename device.");
    }
    setSaving(false);
  };

  if (!editing) {
    return (
      <div className="flex items-center gap-2">
        <h1 className="text-2xl font-bold text-text-primary">{currentName}</h1>
        {canRename && (
          <button
            onClick={() => {
              setName(currentName);
              setEditing(true);
            }}
            className="p-1.5 rounded-md text-text-muted hover:text-primary hover:bg-primary/10 transition-all"
            title="Rename"
          >
            <PencilSquareIcon className="w-4 h-4" />
          </button>
        )}
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center gap-2">
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") void handleSave();
            if (e.key === "Escape") setEditing(false);
          }}
          autoFocus
          className="text-2xl font-bold text-text-primary bg-transparent border-b-2 border-primary/50 focus:outline-none focus:border-primary w-full max-w-md"
        />
        <button
          onClick={() => void handleSave()}
          disabled={saving}
          className="p-1.5 rounded-md text-accent-green hover:bg-accent-green/10 transition-all"
        >
          <CheckIcon className="w-4 h-4" strokeWidth={2} />
        </button>
        <button
          onClick={() => setEditing(false)}
          className="p-1.5 rounded-md text-text-muted hover:bg-hover-medium transition-all"
        >
          <XMarkIcon className="w-4 h-4" strokeWidth={2} />
        </button>
      </div>
      {error && <p className="text-2xs text-accent-red mt-1">{error}</p>}
    </div>
  );
}

/* ─── Custom Fields Section ─── */
function CustomFieldsSection({
  uid,
  customFields,
}: {
  uid: string;
  customFields: Record<string, string>;
}) {
  const mutation = useUpdateDeviceCustomFields();
  const canEdit = useHasPermission("device:rename");
  const [keyInput, setKeyInput] = useState("");
  const [valueInput, setValueInput] = useState("");
  const [adding, setAdding] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [confirmKey, setConfirmKey] = useState<string | null>(null);

  const handleAdd = async () => {
    const key = keyInput.trim();
    const value = valueInput.trim();
    if (!key || !value) return;
    if (key in customFields) {
      setError("This key already exists.");
      return;
    }
    setError(null);
    setAdding(true);
    try {
      await mutation.mutateAsync({
        path: { uid },
        body: { name: "", custom_fields: { ...customFields, [key]: value } },
      });
      setKeyInput("");
      setValueInput("");
    } catch {
      setError("Failed to add custom field.");
    }
    setAdding(false);
  };

  const handleRemove = async (key: string) => {
    const updated = { ...customFields };
    delete updated[key];
    try {
      await mutation.mutateAsync({
        path: { uid },
        body: { name: "", custom_fields: updated },
      });
    } catch {
      /* invalidation handles UI update */
    }
  };

  return (
    <div>
      <h3 className={LABEL + " mb-3"}>Custom Fields</h3>
      <dl className="space-y-2 mb-3">
        {Object.entries(customFields).map(([key, value]) => (
          <div key={key} className="flex items-center justify-between gap-2">
            <div className="flex items-center gap-2 min-w-0">
              <span className="text-xs font-mono text-text-muted shrink-0">{key}:</span>
              <span className="text-sm text-text-primary font-medium truncate">{value}</span>
            </div>
            {canEdit && (
              confirmKey === key
                ? (
                  <div className="flex items-center gap-1 shrink-0">
                    <span className="text-2xs text-text-muted">Remove?</span>
                    <button
                      onClick={() => { void handleRemove(key); setConfirmKey(null); }}
                      className="px-1.5 py-0.5 rounded text-2xs font-semibold bg-accent-red/90 hover:bg-accent-red text-white transition-all"
                    >
                      Yes
                    </button>
                    <button
                      onClick={() => setConfirmKey(null)}
                      className="px-1.5 py-0.5 rounded text-2xs font-semibold text-text-muted hover:text-text-primary hover:bg-hover-subtle transition-all"
                    >
                      No
                    </button>
                  </div>
                )
                : (
                  <button
                    onClick={() => setConfirmKey(key)}
                    className="shrink-0 p-1 rounded-md text-text-muted hover:text-accent-red hover:bg-accent-red/10 transition-all"
                  >
                    <XMarkIcon className="w-3 h-3" strokeWidth={2} />
                  </button>
                )
            )}
          </div>
        ))}
      </dl>
      {canEdit && (
        <div className="flex items-center gap-1.5">
          <input
            type="text"
            value={keyInput}
            onChange={(e) => { setKeyInput(e.target.value); setError(null); }}
            onKeyDown={(e) => { if (e.key === "Enter") { e.preventDefault(); void handleAdd(); } }}
            placeholder="key"
            className="w-24 px-2.5 py-1 bg-card border border-border rounded-md text-xs text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-primary/40 transition-all"
          />
          <span className="text-text-muted text-xs">:</span>
          <input
            type="text"
            value={valueInput}
            onChange={(e) => { setValueInput(e.target.value); setError(null); }}
            onKeyDown={(e) => { if (e.key === "Enter") { e.preventDefault(); void handleAdd(); } }}
            placeholder="value"
            className="w-32 px-2.5 py-1 bg-card border border-border rounded-md text-xs text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-primary/40 transition-all"
          />
          <button
            onClick={() => void handleAdd()}
            disabled={adding || !keyInput.trim() || !valueInput.trim()}
            className="p-1 rounded-md text-text-muted hover:text-primary hover:bg-primary/10 disabled:opacity-soft transition-all"
          >
            <PlusIcon className="w-4 h-4" strokeWidth={2} />
          </button>
        </div>
      )}
      {error && <p className="text-2xs text-accent-red mt-1.5">{error}</p>}
    </div>
  );
}

/* ─── Page ─── */
export default function DeviceDetails() {
  const { uid } = useParams<{ uid: string }>();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { device, isLoading } = useDevice(uid ?? "");
  const removeMutation = useRemoveDevice();
  const tenantId = useAuthStore((s) => s.tenant) ?? "";
  const { namespace: currentNamespace } = useNamespace(tenantId);
  const existingSession = useTerminalStore((s) =>
    s.sessions.find((sess) => sess.deviceUid === uid),
  );
  const restoreTerminal = useTerminalStore((s) => s.restore);
  const [connectOpen, setConnectOpen] = useState(false);
  const [showDelete, setShowDelete] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [operation, setOperation] = useState<{
    device: { uid: string; name: string };
    action: "accept" | "reject" | "remove";
  } | null>(null);

  // Auto-open connect drawer if ?connect=true (adjust during render)
  const shouldAutoConnect
    = searchParams.get("connect") === "true"
      && device?.online
      && !existingSession;

  const [autoConnectDone, setAutoConnectDone] = useState(false);
  if (shouldAutoConnect && !autoConnectDone) {
    setAutoConnectDone(true);
    setConnectOpen(true);
  }
  if (!shouldAutoConnect && autoConnectDone) {
    setAutoConnectDone(false);
  }

  // Restore existing terminal session (side effect only, no setState)
  useEffect(() => {
    if (
      searchParams.get("connect") === "true"
      && device?.online
      && existingSession
    ) {
      restoreTerminal(existingSession.id);
    }
  }, [searchParams, device, existingSession, restoreTerminal]);

  if (isLoading || !device) {
    return (
      <div className="flex items-center justify-center py-24">
        <span className="w-5 h-5 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
      </div>
    );
  }

  const nsName = currentNamespace?.name ?? "";
  const sshid = nsName ? buildSshid(nsName, device.name) : device.uid;

  const tags: string[] = Array.isArray(device.tags)
    ? device.tags.map((t) =>
      typeof t === "object" && t !== null && "name" in t ? t.name : String(t),
    )
    : [];

  const handleDelete = async () => {
    setDeleting(true);
    try {
      await removeMutation.mutateAsync({ path: { uid: device.uid } });
      setShowDelete(false);
      void navigate("/devices");
    } catch {
      setDeleting(false);
    }
  };

  const handleDeviceActionSuccess = () => {
    if (!operation) return;

    if (operation.action === "remove") void navigate("/devices");
  };

  return (
    <div className="animate-fade-in">
      {/* Breadcrumb */}
      <div className="flex items-center gap-1.5 mb-5">
        <Link
          to="/devices"
          className="text-2xs font-mono text-text-muted hover:text-primary transition-colors"
        >
          Devices
        </Link>
        <ChevronRightIcon
          className="w-3 h-3 text-text-muted/40"
          strokeWidth={2}
        />
        <span className="text-2xs font-mono text-text-secondary">
          {device.name}
        </span>
      </div>

      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-4 mb-8">
        <div className="flex items-start gap-4">
          {/* Device icon with status */}
          <div className="relative shrink-0">
            <div className="w-14 h-14 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center">
              <CpuChipIcon className="w-7 h-7 text-primary" />
            </div>
            <span
              className={`absolute -bottom-1 -right-1 w-4 h-4 rounded-full border-2 border-background ${
                device.online
                  ? "bg-accent-green shadow-[0_0_8px_rgba(130,165,104,0.5)]"
                  : "bg-text-muted/40"
              }`}
            />
          </div>

          <div>
            <RenameSection uid={device.uid} currentName={device.name} />
            <div className="flex items-center gap-2 mt-1.5">
              <span
                className={`inline-flex items-center gap-1 px-2 py-0.5 text-2xs font-semibold rounded-md ${
                  device.online
                    ? "bg-accent-green/10 text-accent-green border border-accent-green/20"
                    : "bg-text-muted/10 text-text-muted border border-border"
                }`}
              >
                <span
                  className={`w-1.5 h-1.5 rounded-full ${device.online ? "bg-accent-green" : "bg-text-muted/60"}`}
                />
                {device.online ? "Online" : "Offline"}
              </span>
              <span
                className={`inline-flex items-center px-2 py-0.5 text-2xs font-medium rounded-md ${
                  device.status === "accepted"
                    ? "bg-accent-green/10 text-accent-green"
                    : device.status === "pending"
                      ? "bg-accent-yellow/10 text-accent-yellow"
                      : "bg-accent-red/10 text-accent-red"
                }`}
              >
                {device.status.charAt(0).toUpperCase() + device.status.slice(1)}
              </span>
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2 shrink-0">
          {device.status === "accepted" && (
            <>
              <RestrictedAction action="device:connect">
                <button
                  onClick={() => {
                    if (existingSession) {
                      restoreTerminal(existingSession.id);
                    } else {
                      setConnectOpen(true);
                    }
                  }}
                  disabled={!device.online}
                  className="flex items-center gap-2 px-4 py-2.5 bg-accent-green/90 hover:bg-accent-green text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all"
                >
                  <ChevronDoubleRightIcon className="w-4 h-4" strokeWidth={2} />
                  Connect
                </button>
              </RestrictedAction>
              <RestrictedAction action="device:remove">
                <button
                  onClick={() => setShowDelete(true)}
                  className="p-2.5 rounded-lg text-text-muted hover:text-accent-red hover:bg-accent-red/10 border border-border transition-all"
                  title="Delete device"
                >
                  <TrashIcon className="w-4 h-4" />
                </button>
              </RestrictedAction>
            </>
          )}
          {device.status === "pending" && (
            <>
              <RestrictedAction action="device:accept">
                <button
                  onClick={() => setOperation({ device, action: "accept" })}
                  className="flex items-center gap-2 px-4 py-2.5 bg-accent-green/90 hover:bg-accent-green text-white rounded-lg text-sm font-semibold transition-all"
                >
                  Accept
                </button>
              </RestrictedAction>
              <RestrictedAction action="device:reject">
                <button
                  onClick={() => setOperation({ device, action: "reject" })}
                  className="flex items-center gap-2 px-4 py-2.5 bg-accent-yellow/90 hover:bg-accent-yellow text-white rounded-lg text-sm font-semibold transition-all"
                >
                  Reject
                </button>
              </RestrictedAction>
            </>
          )}
          {device.status === "rejected" && (
            <>
              <RestrictedAction action="device:accept">
                <button
                  onClick={() => setOperation({ device, action: "accept" })}
                  className="flex items-center gap-2 px-4 py-2.5 bg-accent-green/90 hover:bg-accent-green text-white rounded-lg text-sm font-semibold transition-all"
                >
                  Accept
                </button>
              </RestrictedAction>
              <RestrictedAction action="device:remove">
                <button
                  onClick={() => setOperation({ device, action: "remove" })}
                  className="flex items-center gap-2 px-4 py-2.5 bg-accent-red/90 hover:bg-accent-red text-white rounded-lg text-sm font-semibold transition-all"
                >
                  Remove
                </button>
              </RestrictedAction>
            </>
          )}
        </div>
      </div>

      {/* SSHID Banner */}
      {device.status === "accepted" && (
        <div className="bg-card border border-border rounded-xl p-4 mb-6 flex items-center justify-between gap-4">
          <div>
            <p className={LABEL}>SSHID</p>
            <code className="text-sm font-mono text-accent-cyan mt-0.5 block">
              {sshid}
            </code>
          </div>
          <CopyButton text={sshid} />
        </div>
      )}

      {/* Info Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
        <div className="bg-card border border-border rounded-xl p-5 space-y-4">
          <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2">
            <InformationCircleIcon className="w-4 h-4 text-primary" />
            Identity
          </h3>
          <dl className="space-y-3">
            <InfoItem
              label="UID"
              value={device.uid}
              mono
              copyable
              truncate={8}
            />
            <InfoItem
              label="MAC Address"
              value={device.identity?.mac ?? ""}
              mono
              copyable
            />
            <InfoItem label="Remote Address" value={device.remote_addr ?? ""} mono />
          </dl>
        </div>

        <div className="bg-card border border-border rounded-xl p-5 space-y-4">
          <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2">
            <ComputerDesktopIcon className="w-4 h-4 text-primary" />
            System
          </h3>
          <dl className="space-y-3">
            <InfoItem
              label="Operating System"
              value={device.info?.pretty_name ?? ""}
            />
            <InfoItem label="Architecture" value={device.info?.arch ?? ""} mono />
            <div>
              <dt className={LABEL}>Platform</dt>
              <dd className="mt-1">
                {device.info?.platform
                  ? (
                    <PlatformBadge platform={device.info.platform} />
                  )
                  : (
                    <span className="text-sm text-text-muted">—</span>
                  )}
              </dd>
            </div>
            <InfoItem label="Agent Version" value={device.info?.version ?? ""} mono />
          </dl>
        </div>

        <div className="bg-card border border-border rounded-xl p-5 space-y-4">
          <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2">
            <ClockIcon className="w-4 h-4 text-primary" />
            Timeline
          </h3>
          <dl className="space-y-3">
            <div>
              <dt className={LABEL}>Created</dt>
              <dd className={VALUE}>{formatDateFull(device.created_at)}</dd>
            </div>
            <div>
              <dt className={LABEL}>Last Seen</dt>
              <dd className="flex items-center gap-2 mt-0.5">
                <span className="text-sm text-text-primary font-medium">
                  {formatRelative(device.last_seen)}
                </span>
                <span className="text-2xs text-text-muted">
                  {formatDateFull(device.last_seen)}
                </span>
              </dd>
            </div>
            <div>
              <dt className={LABEL}>Status Updated</dt>
              <dd className={VALUE}>
                {formatDateFull(device.status_update_at ?? "")}
              </dd>
            </div>
          </dl>
        </div>
      </div>

      {/* Tags + Custom Fields */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
        <div className="bg-card border border-border rounded-xl p-5">
          <TagsSection uid={device.uid} tags={tags} />
        </div>
        <div className="bg-card border border-border rounded-xl p-5">
          <CustomFieldsSection
            uid={device.uid}
            customFields={device.custom_fields ?? {}}
          />
        </div>
      </div>

      {/* Delete Dialog */}
      {showDelete && (
        <div className="fixed inset-0 z-[70] flex items-center justify-center">
          <div
            className="absolute inset-0 bg-black/60 backdrop-blur-sm"
            onClick={() => setShowDelete(false)}
          />
          <div className="relative bg-surface border border-border rounded-2xl w-full max-w-sm mx-4 p-6 shadow-2xl animate-slide-up">
            <h2 className="text-base font-semibold text-text-primary mb-2">
              Delete Device
            </h2>
            <p className="text-sm text-text-muted mb-6">
              Are you sure you want to delete
              {" "}
              <span className="font-medium text-text-primary">
                {device.name}
              </span>
              ? This action cannot be undone.
            </p>
            <div className="flex justify-end gap-2">
              <button
                onClick={() => setShowDelete(false)}
                className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={() => void handleDelete()}
                disabled={deleting}
                className="px-5 py-2.5 bg-accent-red/90 hover:bg-accent-red text-white rounded-lg text-sm font-semibold disabled:opacity-dim transition-all"
              >
                {deleting ? "Deleting..." : "Delete"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Connect Drawer */}
      <ConnectDrawer
        open={connectOpen}
        onClose={() => setConnectOpen(false)}
        deviceUid={device.uid}
        deviceName={device.name}
        sshid={sshid}
      />

      {/* Action Dialog (accept/reject/remove for pending/rejected devices) */}
      <DeviceActionDialog
        key={operation ? `${operation.action}/${operation.device.uid}` : "closed"}
        open={!!operation}
        device={operation?.device ?? null}
        action={operation?.action ?? "accept"}
        onClose={() => setOperation(null)}
        onSuccess={handleDeviceActionSuccess}
      />
    </div>
  );
}
