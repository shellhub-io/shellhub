import { useEffect, useState, FormEvent } from "react";
import { Link } from "react-router-dom";
import {
  CheckIcon,
  ChatBubbleBottomCenterTextIcon,
  PencilSquareIcon,
  ChevronDownIcon,
  Cog6ToothIcon,
  PencilIcon,
  TagIcon,
  FingerPrintIcon,
  VideoCameraIcon,
  TrashIcon,
  ArrowRightStartOnRectangleIcon,
} from "@heroicons/react/24/outline";
import { useNamespacesStore } from "../stores/namespacesStore";
import { useAuthStore } from "../stores/authStore";
import PageHeader from "../components/common/PageHeader";
import CopyButton from "../components/common/CopyButton";
import Drawer from "../components/common/Drawer";
import ConfirmDialog from "../components/common/ConfirmDialog";
import { LABEL, INPUT } from "../utils/styles";

const NAME_REGEX = /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/;

function validateName(name: string): string | null {
  if (name.length < 3) return "Name must be at least 3 characters";
  if (name.length > 30) return "Name must be at most 30 characters";
  if (name.includes(".")) return "Name cannot contain dots";
  if (!NAME_REGEX.test(name))
    return "Only lowercase letters, numbers, and hyphens allowed";
  return null;
}

/* ─── Settings Card ─── */

function SettingsCard({
  title,
  children,
  danger,
}: {
  title: string;
  children: React.ReactNode;
  danger?: boolean;
}) {
  return (
    <div
      className={`bg-card border rounded-xl overflow-hidden ${danger ? "border-accent-red/20 border-l-2 border-l-accent-red/40" : "border-border"}`}
    >
      <div
        className={`px-5 py-3.5 border-b ${danger ? "border-accent-red/10" : "border-border"}`}
      >
        <h3
          className={`text-sm font-semibold ${danger ? "text-accent-red" : "text-text-primary"}`}
        >
          {title}
        </h3>
      </div>
      <div className="divide-y divide-border">{children}</div>
    </div>
  );
}

/* ─── Settings Row ─── */

function SettingsRow({
  icon,
  title,
  description,
  children,
}: {
  icon: React.ReactNode;
  title: string;
  description: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex items-center justify-between gap-6 px-5 py-4">
      <div className="flex items-start gap-3 min-w-0 flex-1">
        <span className="w-8 h-8 rounded-lg bg-hover-medium border border-border flex items-center justify-center text-text-muted shrink-0 mt-0.5">
          {icon}
        </span>
        <div className="min-w-0">
          <p className="text-sm font-medium text-text-primary">{title}</p>
          <p className="text-2xs text-text-muted mt-0.5 leading-relaxed">
            {description}
          </p>
        </div>
      </div>
      <div className="shrink-0">{children}</div>
    </div>
  );
}

/* ─── Edit Name Drawer ─── */

function EditNameDrawer({
  open,
  onClose,
  currentName,
  tenantId,
}: {
  open: boolean;
  onClose: () => void;
  currentName: string;
  tenantId: string;
}) {
  const updateNamespace = useNamespacesStore((s) => s.updateNamespace);
  const [name, setName] = useState(currentName);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (open) {
      setName(currentName);
      setError("");
    }
  }, [open, currentName]);

  const validationError = name !== currentName ? validateName(name) : null;

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    const err = validateName(name);
    if (err) {
      setError(err);
      return;
    }
    setSubmitting(true);
    setError("");
    try {
      await updateNamespace(tenantId, { name });
      onClose();
    } catch {
      setError("Failed to rename namespace. The name may already be taken.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Rename Namespace"
      bodyClassName="flex-1 overflow-y-auto px-6 py-5"
      footer={
        <>
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleSubmit}
            disabled={
              !name.trim() ||
              name === currentName ||
              !!validationError ||
              submitting
            }
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
          >
            {submitting ? (
              <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
            ) : (
              <CheckIcon className="w-4 h-4" strokeWidth={2} />
            )}
            Save
          </button>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="space-y-5">
        <div>
          <label className={LABEL}>Namespace Name</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value.toLowerCase())}
            autoFocus={open}
            className={INPUT}
          />
          <p className="text-2xs text-text-muted mt-1.5">
            3-30 characters, lowercase letters, numbers, and hyphens. No dots.
          </p>
          {validationError && (
            <p className="text-2xs text-accent-red mt-1">{validationError}</p>
          )}
        </div>
        {error && <p className="text-2xs text-accent-red">{error}</p>}
      </form>
    </Drawer>
  );
}

/* ─── Delete Namespace Dialog ─── */

function DeleteDialog({
  namespaceName,
  tenantId,
  onClose,
}: {
  namespaceName: string;
  tenantId: string;
  onClose: () => void;
}) {
  const deleteNs = useNamespacesStore((s) => s.deleteNamespace);
  const [confirm, setConfirm] = useState("");
  const [error, setError] = useState("");

  const canDelete = confirm === namespaceName;

  return (
    <ConfirmDialog
      open
      onClose={onClose}
      onConfirm={async () => {
        setError("");
        try {
          await deleteNs(tenantId);
        } catch {
          setError("Failed to delete namespace.");
          throw new Error();
        }
      }}
      title="Delete Namespace"
      description={
        <>
          This action is{" "}
          <span className="font-medium text-accent-red">permanent</span> and
          cannot be undone. All devices, sessions, and data will be lost.
        </>
      }
      confirmLabel="Delete Namespace"
      confirmDisabled={!canDelete}
    >
      <div className="mb-4">
        <label className={LABEL}>
          Type &ldquo;{namespaceName}&rdquo; to confirm
        </label>
        <input
          type="text"
          value={confirm}
          onChange={(e) => setConfirm(e.target.value)}
          placeholder={namespaceName}
          className={INPUT}
          autoFocus
        />
      </div>
      {error && <p className="text-2xs text-accent-red mb-3">{error}</p>}
    </ConfirmDialog>
  );
}

/* ─── Leave Namespace Dialog ─── */

function LeaveDialog({
  tenantId,
  onClose,
}: {
  tenantId: string;
  onClose: () => void;
}) {
  const leaveNs = useNamespacesStore((s) => s.leaveNamespace);
  const [error, setError] = useState("");

  return (
    <ConfirmDialog
      open
      onClose={onClose}
      onConfirm={async () => {
        setError("");
        try {
          await leaveNs(tenantId);
        } catch {
          setError("Failed to leave namespace.");
          throw new Error();
        }
      }}
      title="Leave Namespace"
      description="You will lose access to all devices and sessions. To rejoin, someone will need to invite you again."
      confirmLabel="Leave"
    >
      {error && <p className="text-2xs text-accent-red mb-3">{error}</p>}
    </ConfirmDialog>
  );
}

/* ─── Banner Collapsible ─── */

function BannerPreview({
  banner,
  canEdit,
}: {
  banner: string;
  canEdit: boolean;
}) {
  const [open, setOpen] = useState(false);

  if (!banner) {
    return (
      <SettingsRow
        icon={<ChatBubbleBottomCenterTextIcon className="w-4 h-4" />}
        title="SSH Banner"
        description="Message shown when users connect via SSH"
      >
        {canEdit && (
          <Link
            to="/settings/banner"
            className="inline-flex p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-hover-medium transition-colors"
            title="Edit"
          >
            <PencilSquareIcon className="w-4 h-4" />
          </Link>
        )}
      </SettingsRow>
    );
  }

  return (
    <div className="px-5 py-4">
      {/* Header */}
      <div className="flex items-center justify-between gap-4">
        <button
          type="button"
          onClick={() => setOpen((v) => !v)}
          className="flex items-start gap-3 min-w-0 flex-1 text-left"
        >
          <span className="w-8 h-8 rounded-lg bg-hover-medium border border-border flex items-center justify-center text-text-muted shrink-0 mt-0.5">
            <ChatBubbleBottomCenterTextIcon className="w-4 h-4" />
          </span>
          <div className="min-w-0 pt-0.5">
            <p className="text-sm font-medium text-text-primary">SSH Banner</p>
            <p className="text-2xs text-text-muted mt-0.5">
              Message shown when users connect via SSH
            </p>
          </div>
        </button>
        <div className="flex items-center gap-1 shrink-0">
          {canEdit && (
            <Link
              to="/settings/banner"
              className="inline-flex p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-hover-medium transition-colors"
              title="Edit"
            >
              <PencilSquareIcon className="w-4 h-4" />
            </Link>
          )}
          <button
            type="button"
            onClick={() => setOpen((v) => !v)}
            className="inline-flex p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-hover-medium transition-colors"
          >
            <ChevronDownIcon
              className={`w-4 h-4 transition-transform duration-200 ${open ? "rotate-180" : ""}`}
            />
          </button>
        </div>
      </div>

      {/* Collapsible content */}
      <div className="ml-11 mt-3">
        <div
          className={`relative overflow-hidden rounded-lg border border-border bg-card transition-all duration-200 ease-out ${open ? "max-h-[500px]" : "max-h-[120px]"}`}
        >
          <pre className="px-3 py-2.5 text-xs font-mono text-text-secondary leading-relaxed whitespace-pre-wrap break-words">
            {banner}
          </pre>
          {!open && (
            <>
              <div className="absolute inset-x-0 bottom-0 h-16 bg-gradient-to-t from-card via-card/70 to-transparent pointer-events-none" />
              <button
                type="button"
                onClick={() => setOpen(true)}
                className="absolute inset-x-0 bottom-0 flex items-center justify-center pb-2.5 pt-4"
              >
                <span className="inline-flex items-center gap-1 px-3 py-1 text-2xs font-semibold text-text-primary bg-surface border border-border-light rounded-full shadow-sm hover:bg-card hover:border-primary/30 transition-all">
                  Show more
                  <ChevronDownIcon
                    className="w-3 h-3 animate-bounce-subtle"
                    strokeWidth={2}
                  />
                </span>
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  );
}

/* ─── Page ─── */

export default function Settings() {
  const { userId, tenant: tenantId, role: sessionRole } = useAuthStore();
  const { currentNamespace, fetchCurrent, updateNamespace } =
    useNamespacesStore();
  const [editNameOpen, setEditNameOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [leaveOpen, setLeaveOpen] = useState(false);
  const [togglingRecord, setTogglingRecord] = useState(false);

  useEffect(() => {
    if (tenantId) fetchCurrent(tenantId);
  }, [tenantId, fetchCurrent]);

  const ns = currentNamespace;
  const isOwner = ns?.owner === userId;
  const currentMember = ns?.members?.find((m) => m.id === userId);
  const role =
    currentMember?.role ?? (isOwner ? "owner" : (sessionRole ?? "observer"));
  const canEdit = isOwner || role === "administrator";
  const settings = ns?.settings;
  const sessionRecord = settings?.session_record ?? false;
  const banner = settings?.connection_announcement ?? "";

  const handleToggleRecord = async () => {
    if (!tenantId || togglingRecord) return;
    setTogglingRecord(true);
    try {
      await updateNamespace(tenantId, {
        settings: { session_record: !sessionRecord },
      });
    } catch {
      /* state didn't change */
    } finally {
      setTogglingRecord(false);
    }
  };

  if (!ns) {
    return (
      <div className="flex items-center justify-center py-32">
        <span className="w-5 h-5 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
      </div>
    );
  }

  return (
    <div>
      <PageHeader
        icon={<Cog6ToothIcon className="w-6 h-6" />}
        overline="Namespace"
        title="Settings"
        description="Manage namespace configuration, SSH options, and access"
      />

      <div className="space-y-6 animate-fade-in">
        {/* ── General ── */}
        <SettingsCard title="General">
          <SettingsRow
            icon={<PencilIcon className="w-4 h-4" />}
            title="Name"
            description="The display name for this namespace"
          >
            <div className="flex items-center gap-2">
              <span className="text-sm font-mono text-text-secondary">
                {ns.name}
              </span>
              {canEdit && (
                <button
                  onClick={() => setEditNameOpen(true)}
                  className="p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-hover-medium transition-colors"
                  title="Rename"
                >
                  <PencilSquareIcon className="w-4 h-4" />
                </button>
              )}
            </div>
          </SettingsRow>

          <SettingsRow
            icon={<TagIcon className="w-4 h-4" />}
            title="Type"
            description="Defines whether this namespace belongs to one user or a team"
          >
            <span
              className={`inline-flex items-center px-2.5 py-1 text-2xs font-mono font-semibold rounded border ${
                ns.type === "team"
                  ? "bg-primary/10 text-primary border-primary/20"
                  : "bg-accent-yellow/10 text-accent-yellow border-accent-yellow/20"
              }`}
            >
              {ns.type ?? "personal"}
            </span>
          </SettingsRow>

          <SettingsRow
            icon={<FingerPrintIcon className="w-4 h-4" />}
            title="Tenant ID"
            description="Use this identifier when integrating with the ShellHub API"
          >
            <div className="flex items-center gap-1">
              <code className="text-xs font-mono text-accent-cyan select-all">
                {ns.tenant_id}
              </code>
              <CopyButton text={ns.tenant_id} size="md" />
            </div>
          </SettingsRow>
        </SettingsCard>

        {/* ── SSH ── */}
        <SettingsCard title="SSH">
          {/* Session Recording */}
          <SettingsRow
            icon={<VideoCameraIcon className="w-4 h-4" />}
            title="Session Recording"
            description="Record SSH sessions for audit and playback"
          >
            <div
              className={`inline-flex items-center h-7 bg-card border border-border rounded-md p-0.5 ${!canEdit || togglingRecord ? "opacity-40 pointer-events-none" : ""}`}
            >
              <button
                type="button"
                onClick={() => {
                  if (sessionRecord) handleToggleRecord();
                }}
                className={`h-full px-2.5 text-2xs font-medium rounded transition-all duration-150 ${
                  !sessionRecord
                    ? "bg-hover-strong text-text-secondary border border-border-light"
                    : "text-text-muted hover:text-text-secondary border border-transparent"
                }`}
              >
                Off
              </button>
              <button
                type="button"
                onClick={() => {
                  if (!sessionRecord) handleToggleRecord();
                }}
                className={`h-full px-2.5 text-2xs font-medium rounded transition-all duration-150 ${
                  sessionRecord
                    ? "bg-primary/15 text-primary border border-primary/25"
                    : "text-text-muted hover:text-text-secondary border border-transparent"
                }`}
              >
                On
              </button>
            </div>
          </SettingsRow>

          {/* SSH Banner */}
          <BannerPreview banner={banner} canEdit={canEdit} />
        </SettingsCard>

        {/* ── Danger Zone ── */}
        <SettingsCard title="Danger Zone" danger>
          {isOwner ? (
            <SettingsRow
              icon={<TrashIcon className="w-4 h-4 text-accent-red" />}
              title="Delete Namespace"
              description="Permanently removes all devices, sessions, keys, and configuration. This cannot be undone."
            >
              <button
                onClick={() => setDeleteOpen(true)}
                className="px-4 py-2 bg-accent-red/10 hover:bg-accent-red/20 text-accent-red border border-accent-red/20 rounded-lg text-sm font-semibold transition-all"
              >
                Delete
              </button>
            </SettingsRow>
          ) : (
            <SettingsRow
              icon={
                <ArrowRightStartOnRectangleIcon className="w-4 h-4 text-accent-red" />
              }
              title="Leave Namespace"
              description="You will lose access immediately. To rejoin, someone will need to invite you again."
            >
              <button
                onClick={() => setLeaveOpen(true)}
                className="px-4 py-2 bg-accent-red/10 hover:bg-accent-red/20 text-accent-red border border-accent-red/20 rounded-lg text-sm font-semibold transition-all"
              >
                Leave
              </button>
            </SettingsRow>
          )}
        </SettingsCard>
      </div>

      {/* Drawers & Dialogs */}
      <EditNameDrawer
        open={editNameOpen}
        onClose={() => setEditNameOpen(false)}
        currentName={ns.name}
        tenantId={ns.tenant_id}
      />
      {deleteOpen && (
        <DeleteDialog
          namespaceName={ns.name}
          tenantId={ns.tenant_id}
          onClose={() => setDeleteOpen(false)}
        />
      )}
      {leaveOpen && (
        <LeaveDialog
          tenantId={ns.tenant_id}
          onClose={() => setLeaveOpen(false)}
        />
      )}
    </div>
  );
}
