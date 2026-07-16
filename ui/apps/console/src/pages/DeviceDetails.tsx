import { useEffect, useState } from "react";
import { useParams, useNavigate, useSearchParams } from "react-router-dom";
import Breadcrumb from "@/components/common/Breadcrumb";
import { LABEL_BASE } from "@/utils/styles";
import {
  TrashIcon,
  ComputerDesktopIcon,
  CpuChipIcon,
  ChevronDoubleRightIcon,
} from "@heroicons/react/24/outline";
import { useDevice } from "../hooks/useDevice";
import { useDeviceActions } from "../hooks/useDeviceActions";
import {
  useRenameDevice,
  useAddDeviceTag,
  useRemoveDeviceTag,
} from "../hooks/useDeviceMutations";
import { useNamespace } from "../hooks/useNamespaces";
import { useInstallKeys } from "../hooks/useInstallKeys";
import { resolveEnrollmentSource } from "@/pages/install-keys/helpers";
import { DeprecatedBadge } from "@/pages/install-keys/constants";
import { useAuthStore } from "../stores/authStore";
import { useTerminalStore } from "../stores/terminalStore";
import DeviceActionsPortal from "./devices/DeviceActionsPortal";
import ConnectDrawer from "../components/ConnectDrawer";
import CopyButton from "../components/common/CopyButton";
import PlatformBadge from "../components/common/PlatformBadge";
import { buildSshid } from "../utils/sshid";
import RestrictedAction from "../components/common/RestrictedAction";
import PageLoader from "@/components/common/PageLoader";
import IdentityCard from "@/components/common/IdentityCard";
import InfoItem from "@/components/common/InfoItem";
import TimelineCard from "@/components/common/TimelineCard";
import TagsSection from "@/components/common/TagsSection";
import RenameSection from "@/components/common/RenameSection";
import { useHasPermission } from "@/hooks/useHasPermission";
import CustomFieldsSection from "./devices/CustomFieldsSection";
import { Button, Card, IconButton } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";

export default function DeviceDetails() {
  const { uid } = useParams<{ uid: string }>();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { device, isLoading } = useDevice(uid ?? "");
  const tenantId = useAuthStore((s) => s.tenant) ?? "";
  const { namespace: currentNamespace } = useNamespace(tenantId);
  const { installKeys } = useInstallKeys({ perPage: 100 });
  const existingSession = useTerminalStore((s) =>
    s.sessions.find((sess) => sess.deviceUid === uid),
  );
  const restoreTerminal = useTerminalStore((s) => s.restore);
  const [connectOpen, setConnectOpen] = useState(false);
  const renameMutation = useRenameDevice();
  const canRename = useHasPermission("device:rename");
  const addTagMutation = useAddDeviceTag();
  const removeTagMutation = useRemoveDeviceTag();
  const actionsController = useDeviceActions({
    onSuccess: (action) => {
      if (action === "remove") void navigate("/devices");
    },
  });

  // Auto-open connect drawer if ?connect=true (adjust during render)
  const shouldAutoConnect =
    searchParams.get("connect") === "true" &&
    device?.online &&
    !existingSession;

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
      searchParams.get("connect") === "true" &&
      device?.online &&
      existingSession
    ) {
      restoreTerminal(existingSession.id);
    }
  }, [searchParams, device, existingSession, restoreTerminal]);

  if (isLoading || !device) {
    return <PageLoader label="Loading device details" />;
  }

  const nsName = currentNamespace?.name ?? "";
  const sshid = nsName ? buildSshid(nsName, device.name) : device.uid;

  const enrollment = resolveEnrollmentSource(
    device.install_key_id,
    installKeys,
  );

  const tags: string[] = Array.isArray(device.tags)
    ? device.tags.map((t) =>
        typeof t === "object" && t !== null && "name" in t ? t.name : String(t),
      )
    : [];

  const statusColor =
    device.status === "accepted"
      ? "bg-accent-green/10 text-accent-green"
      : device.status === "pending"
        ? "bg-accent-yellow/10 text-accent-yellow"
        : "bg-accent-red/10 text-accent-red";

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
              className={cn(
                "absolute -bottom-1 -right-1 w-4 h-4 rounded-full border-2 border-background",
                device.online
                  ? "bg-accent-green shadow-[0_0_8px_rgba(130,165,104,0.5)]"
                  : "bg-text-muted/40",
              )}
            />
          </div>

          <div>
            <RenameSection
              uid={device.uid}
              currentName={device.name}
              rename={renameMutation.mutateAsync}
              entityLabel="device"
              canRename={canRename}
            />
            <div className="flex items-center gap-2 mt-1.5">
              <span
                className={cn(
                  "inline-flex items-center gap-1 px-2 py-0.5 text-2xs font-semibold rounded-md",
                  device.online
                    ? "bg-accent-green/10 text-accent-green border border-accent-green/20"
                    : "bg-text-muted/10 text-text-muted border border-border",
                )}
              >
                <span
                  className={cn(
                    "w-1.5 h-1.5 rounded-full",
                    device.online ? "bg-accent-green" : "bg-text-muted/60",
                  )}
                />
                {device.online ? "Online" : "Offline"}
              </span>
              <span
                className={cn(
                  "inline-flex items-center px-2 py-0.5 text-2xs font-medium rounded-md",
                  statusColor,
                )}
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
                <Button
                  variant="success"
                  onClick={() => {
                    if (existingSession) {
                      restoreTerminal(existingSession.id);
                    } else {
                      setConnectOpen(true);
                    }
                  }}
                  disabled={!device.online}
                  icon={
                    <ChevronDoubleRightIcon
                      className="w-4 h-4"
                      strokeWidth={2}
                    />
                  }
                >
                  Connect
                </Button>
              </RestrictedAction>
              <RestrictedAction action="device:remove">
                <IconButton
                  variant="danger"
                  size="lg"
                  type="button"
                  title="Delete device"
                  aria-label="Delete device"
                  className="border border-border"
                  onClick={() =>
                    actionsController.requestAction(device, "remove")
                  }
                >
                  <TrashIcon className="w-4 h-4" />
                </IconButton>
              </RestrictedAction>
            </>
          )}
          {device.status === "pending" && (
            <>
              <RestrictedAction action="device:accept">
                <Button
                  variant="success"
                  onClick={() =>
                    actionsController.requestAction(device, "accept")
                  }
                >
                  Accept
                </Button>
              </RestrictedAction>
              <RestrictedAction action="device:reject">
                <Button
                  variant="warning"
                  onClick={() =>
                    actionsController.requestAction(device, "reject")
                  }
                >
                  Reject
                </Button>
              </RestrictedAction>
            </>
          )}
          {device.status === "rejected" && (
            <>
              <RestrictedAction action="device:accept">
                <Button
                  variant="success"
                  onClick={() =>
                    actionsController.requestAction(device, "accept")
                  }
                >
                  Accept
                </Button>
              </RestrictedAction>
              <RestrictedAction action="device:remove">
                <Button
                  variant="destructive"
                  onClick={() =>
                    actionsController.requestAction(device, "remove")
                  }
                >
                  Remove
                </Button>
              </RestrictedAction>
            </>
          )}
        </div>
      </div>

      {/* SSHID Banner */}
      {device.status === "accepted" && (
        <Card className="p-4 mb-6 flex items-center justify-between gap-4">
          <div>
            <p className={LABEL_BASE}>SSHID</p>
            <code className="text-sm font-mono text-accent-cyan mt-0.5 block">
              {sshid}
            </code>
          </div>
          <CopyButton text={sshid} />
        </Card>
      )}

      {/* Info Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
        <IdentityCard
          uid={device.uid}
          mac={device.identity?.mac ?? ""}
          remoteAddr={device.remote_addr ?? ""}
          registeredVia={
            enrollment ? (
              enrollment.kind === "legacy" ? (
                <DeprecatedBadge />
              ) : (
                <span className="text-sm font-medium text-text-primary">
                  {enrollment.name}
                </span>
              )
            ) : (
              <span className="text-sm text-text-muted">—</span>
            )
          }
        />

        <Card className="p-5 space-y-4">
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
            <InfoItem label="Platform">
              {device.info?.platform ? (
                <PlatformBadge platform={device.info.platform} />
              ) : (
                <span className="text-sm text-text-muted">—</span>
              )}
            </InfoItem>
            <InfoItem
              label="Agent Version"
              value={device.info?.version ?? ""}
              mono
            />
          </dl>
        </Card>

        <TimelineCard
          createdAt={device.created_at}
          lastSeen={device.last_seen}
          statusUpdatedAt={device.status_updated_at ?? ""}
        />
      </div>

      {/* Tags + Custom Fields */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
        <Card className="p-5">
          <TagsSection
            uid={device.uid}
            tags={tags}
            addTag={addTagMutation.mutateAsync}
            removeTag={removeTagMutation.mutateAsync}
          />
        </Card>
        <Card className="p-5">
          <CustomFieldsSection
            uid={device.uid}
            customFields={device.custom_fields ?? {}}
          />
        </Card>
      </div>

      {/* Connect Drawer */}
      <ConnectDrawer
        open={connectOpen}
        onClose={() => setConnectOpen(false)}
        deviceUid={device.uid}
        deviceName={device.name}
        sshid={sshid}
      />

      {/* Action Portal (accept/reject/remove for pending/rejected devices) */}
      <DeviceActionsPortal controller={actionsController} />
    </div>
  );
}
