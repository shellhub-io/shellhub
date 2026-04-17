import { useEffect, useState } from "react";
import { useParams, useNavigate, useSearchParams } from "react-router-dom";
import Breadcrumb from "@/components/common/Breadcrumb";
import {
  TrashIcon,
  InformationCircleIcon,
  ComputerDesktopIcon,
  ClockIcon,
  CpuChipIcon,
  ChevronDoubleRightIcon,
  LockOpenIcon,
  LockClosedIcon,
} from "@heroicons/react/24/outline";
import { useDevice } from "../hooks/useDevice";
import { useRemoveDevice, useUpdateDeviceSSH } from "../hooks/useDeviceMutations";
import { useHasPermission } from "../hooks/useHasPermission";
import { useNamespace } from "../hooks/useNamespaces";
import { useAuthStore } from "../stores/authStore";
import { useTerminalStore } from "../stores/terminalStore";
import DeviceActionDialog from "./devices/DeviceActionDialog";
import BillingWarning from "../components/billing/BillingWarning";
import ConnectDrawer from "../components/ConnectDrawer";
import ConfirmDialog from "../components/common/ConfirmDialog";
import CopyButton from "../components/common/CopyButton";
import PlatformBadge from "../components/common/PlatformBadge";
import SettingToggle from "../components/common/SettingToggle";
import { formatDateFull, formatRelative } from "../utils/date";
import { buildSshid } from "../utils/sshid";
import RestrictedAction from "../components/common/RestrictedAction";
import { getConfig } from "../env";
import PageLoader from "@/components/common/PageLoader";
import InfoItem from "./devices/InfoItem";
import TagsSection from "./devices/TagsSection";
import RenameSection from "./devices/RenameSection";
import CustomFieldsSection from "./devices/CustomFieldsSection";
import type { Device } from "../client";

/* ─── Shared styles ─── */
const LABEL =
  "text-2xs font-mono font-semibold uppercase tracking-label text-text-muted";
const VALUE = "text-sm text-text-primary font-medium mt-0.5";
type DeviceSSHSettings = NonNullable<Device["settings"]>;
type DeviceSSHSettingKey = keyof DeviceSSHSettings;

const DEVICE_SSH_SETTINGS: Array<{
  key: DeviceSSHSettingKey;
  title: string;
  description: string;
}> = [
  {
    key: "allow_password",
    title: "Allow Password Authentication",
    description: "Allow SSH connections using password for this device",
  },
  {
    key: "allow_public_key",
    title: "Allow Public Key Authentication",
    description: "Allow SSH connections using public key for this device",
  },
  {
    key: "allow_root",
    title: "Allow Root Login",
    description: "Allow SSH connections as root user for this device",
  },
  {
    key: "allow_empty_passwords",
    title: "Allow Empty Passwords",
    description: "Allow SSH connections with empty passwords for this device",
  },
  {
    key: "allow_tty",
    title: "Allow TTY Allocation",
    description: "Allow terminal (TTY) allocation for this device",
  },
  {
    key: "allow_tcp_forwarding",
    title: "Allow TCP Forwarding",
    description: "Allow TCP port forwarding for this device",
  },
  {
    key: "allow_web_endpoints",
    title: "Allow Web Endpoints",
    description: "Allow HTTP/HTTPS access via ShellHub proxy",
  },
  {
    key: "allow_sftp",
    title: "Allow SFTP",
    description: "Allow SFTP subsystem for this device",
  },
  {
    key: "allow_agent_forwarding",
    title: "Allow Agent Forwarding",
    description: "Allow SSH agent forwarding for this device",
  },
];

function ToggleStateIcon({ enabled }: { enabled: boolean }) {
  return enabled
    ? <LockOpenIcon className="w-4 h-4 text-accent-green" />
    : <LockClosedIcon className="w-4 h-4 text-accent-red" />;
}

/* ─── Page ─── */
export default function DeviceDetails() {
  const { uid } = useParams<{ uid: string }>();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { device, isLoading } = useDevice(uid ?? "");
  const removeMutation = useRemoveDevice();
  const updateSSH = useUpdateDeviceSSH();
  const canUpdateDeviceSettings = useHasPermission("device:update");
  const tenantId = useAuthStore((s) => s.tenant) ?? "";
  const { namespace: currentNamespace } = useNamespace(tenantId);
  const deviceSettings = device?.settings ?? {};
  const existingSession = useTerminalStore((s) =>
    s.sessions.find((sess) => sess.deviceUid === uid),
  );
  const restoreTerminal = useTerminalStore((s) => s.restore);
  const [connectOpen, setConnectOpen] = useState(false);
  const [showDelete, setShowDelete] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const [operation, setOperation] = useState<{
    device: { uid: string; name: string };
    action: "accept" | "reject" | "remove";
  } | null>(null);
  const [billingWarningOpen, setBillingWarningOpen] = useState(false);

  const updateDeviceSetting = async (settings: Partial<DeviceSSHSettings>) => {
    if (!device) {
      return;
    }

    await updateSSH.mutateAsync({
      path: { uid: device.uid },
      body: settings,
    });
  };

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

  useEffect(() => {
    if (
      searchParams.get("connect") === "true" &&
      device?.online &&
      existingSession
    ) {
      restoreTerminal(existingSession.id);
    }
  }, [searchParams, device, existingSession, restoreTerminal]);

  if (isLoading || !device) {
    return (
      <PageLoader label="Loading device details" />
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
    setDeleteError(null);
    try {
      await removeMutation.mutateAsync({ path: { uid: device.uid } });
      setShowDelete(false);
      void navigate("/devices");
    } catch {
      setDeleteError("Failed to delete device. Please try again.");
    }
  };

  const handleDeviceActionSuccess = () => {
    if (!operation) return;
    if (operation.action === "remove") void navigate("/devices");
  };

  return (
    <div className="animate-fade-in">
      <Breadcrumb
        items={[{ label: "Devices", to: "/devices" }, { label: device.name }]}
      />

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
                  type="button"
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
                  type="button"
                  onClick={() => setShowDelete(true)}
                  className="p-2.5 rounded-lg text-text-muted hover:text-accent-red hover:bg-accent-red/10 border border-border transition-all"
                  aria-label="Delete device"
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
                  type="button"
                  onClick={() => setOperation({ device, action: "accept" })}
                  className="flex items-center gap-2 px-4 py-2.5 bg-accent-green/90 hover:bg-accent-green text-white rounded-lg text-sm font-semibold transition-all"
                >
                  Accept
                </button>
              </RestrictedAction>
              <RestrictedAction action="device:reject">
                <button
                  type="button"
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
                  type="button"
                  onClick={() => setOperation({ device, action: "accept" })}
                  className="flex items-center gap-2 px-4 py-2.5 bg-accent-green/90 hover:bg-accent-green text-white rounded-lg text-sm font-semibold transition-all"
                >
                  Accept
                </button>
              </RestrictedAction>
              <RestrictedAction action="device:remove">
                <button
                  type="button"
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
            <InfoItem
              label="Remote Address"
              value={device.remote_addr ?? ""}
              mono
            />
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
            <InfoItem
              label="Architecture"
              value={device.info?.arch ?? ""}
              mono
            />
            <div>
              <dt className={LABEL}>Platform</dt>
              <dd className="mt-1">
                {device.info?.platform ? (
                  <PlatformBadge platform={device.info.platform} />
                ) : (
                  <span className="text-sm text-text-muted">—</span>
                )}
              </dd>
            </div>
            <InfoItem
              label="Agent Version"
              value={device.info?.version ?? ""}
              mono
            />
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

      {/* Settings */}
      <div className="bg-card border border-border rounded-xl p-5 mb-6">
        <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2 mb-4">
          <LockClosedIcon className="w-4 h-4 text-primary" />
          Settings
        </h3>
        <div className="divide-y divide-border -mx-2">
          {DEVICE_SSH_SETTINGS.map((setting) => {
            const enabled = deviceSettings[setting.key] ?? true;

            return (
              <div key={setting.key} className="flex items-center justify-between gap-6 px-2 py-3">
                <div className="flex items-start gap-3 min-w-0 flex-1">
                  <span className="w-8 h-8 rounded-lg bg-hover-medium border border-border flex items-center justify-center text-text-muted shrink-0 mt-0.5">
                    <ToggleStateIcon enabled={enabled} />
                  </span>
                  <div className="min-w-0">
                    <p className="text-sm font-medium text-text-primary">
                      {setting.title}
                    </p>
                    <p className="text-2xs text-text-muted mt-0.5 leading-relaxed">
                      {setting.description}
                    </p>
                  </div>
                </div>
                <div className="shrink-0">
                  <SettingToggle
                    checked={enabled}
                    tone="success"
                    disabled={!canUpdateDeviceSettings || updateSSH.isPending}
                    onChange={(checked) => {
                      return updateDeviceSetting({ [setting.key]: checked });
                    }}
                  />
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Delete Dialog */}
      <ConfirmDialog
        open={showDelete}
        onClose={() => {
          setShowDelete(false);
          setDeleteError(null);
        }}
        onConfirm={handleDelete}
        title="Delete Device"
        description={
          <>
            Are you sure you want to delete{" "}
            <span className="font-medium text-text-primary">{device.name}</span>
            ? This action cannot be undone.
          </>
        }
        confirmLabel="Delete"
        variant="danger"
        errorMessage={deleteError}
      />

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
        key={
          operation ? `${operation.action}/${operation.device.uid}` : "closed"
        }
        open={!!operation}
        device={operation?.device ?? null}
        action={operation?.action ?? "accept"}
        onClose={() => setOperation(null)}
        onSuccess={handleDeviceActionSuccess}
        onBillingWarning={
          getConfig().cloud
            ? () => {
                setOperation(null);
                setBillingWarningOpen(true);
              }
            : undefined
        }
      />
      <BillingWarning
        open={billingWarningOpen}
        onClose={() => setBillingWarningOpen(false)}
      />
    </div>
  );
}
