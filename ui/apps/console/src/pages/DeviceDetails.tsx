import { Link, useNavigate, useParams } from "react-router-dom";
import Breadcrumb from "@/components/common/Breadcrumb";
import {
  TrashIcon,
  ComputerDesktopIcon,
  CpuChipIcon,
  ChevronDoubleRightIcon,
} from "@heroicons/react/24/outline";
import { useActionDialog } from "../hooks/useActionDialog";
import {
  useDevice,
  useRenameDevice,
  useAddDeviceTag,
  useRemoveDeviceTag,
} from "@/hooks/useDeviceRequests";
import { useProvisioningKeys } from "@/hooks/useProvisioningKeyRequests";
import {
  enrollmentSourceName,
  provisioningKeyActivityPath,
  resolveEnrollmentSource,
} from "@/pages/provisioning-keys/helpers";
import { useAuthStore } from "../stores/authStore";
import { useTerminalStore } from "../stores/terminalStore";
import ActionDialog from "@/components/common/ActionDialog";
import { useDeviceActionRunner } from "@/hooks/useDeviceActionRunner";
import PlatformBadge from "../components/common/PlatformBadge";
import SshidCard from "@/components/common/SshidCard";
import { normalizeDeviceTags } from "@/utils/deviceTags";
import RestrictedAction from "../components/common/RestrictedAction";
import PageLoader from "@/components/common/PageLoader";
import IdentityCard from "@/components/common/IdentityCard";
import InfoItem from "@/components/common/InfoItem";
import PairedBy from "@/pages/devices/PairedBy";
import TimelineCard from "@/components/common/TimelineCard";
import TagsSection from "@/components/common/TagsSection";
import RenameSection from "@/components/common/RenameSection";
import { useHasPermission } from "@/hooks/useHasPermission";
import { useConnectFromLink } from "@/hooks/useConnectFromLink";
import CustomFieldsSection from "./devices/CustomFieldsSection";
import { Button, Card, IconButton } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import ResourceNotFound from "@/components/common/ResourceNotFound";

/**
 * Everything about one device: identity, tags, sessions and the actions on it.
 */
export default function DeviceDetails() {
  const { uid } = useParams<{ uid: string }>();
  const navigate = useNavigate();
  const { device, isLoading, error } = useDevice(uid ?? "");
  const tenantId = useAuthStore((s) => s.tenant) ?? "";
  const { provisioningKeys } = useProvisioningKeys();
  const connect = useTerminalStore((s) => s.connect);
  const renameMutation = useRenameDevice();
  const canRename = useHasPermission("device:rename");
  const addTagMutation = useAddDeviceTag();
  const removeTagMutation = useRemoveDeviceTag();
  const actionsController = useActionDialog({
    onSuccess: (operation) => {
      if (operation === "remove") void navigate("/devices");
    },
  });
  const runDeviceAction = useDeviceActionRunner();

  useConnectFromLink(device);

  if (isLoading) {
    return <PageLoader label="Loading device details" />;
  }

  if (error || !device) {
    return (
      <ResourceNotFound
        icon={CpuChipIcon}
        resource="Device"
        backTo="/devices"
      />
    );
  }

  const enrollment = resolveEnrollmentSource(
    device.provisioning_key_id,
    provisioningKeys,
  );

  const { tags } = normalizeDeviceTags(device);

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

      <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-4 mb-8">
        <div className="flex items-start gap-4">
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

        <div className="flex items-center gap-2 shrink-0">
          {device.status === "accepted" && (
            <>
              <RestrictedAction action="device:connect">
                <Button
                  variant="success"
                  onClick={() => connect(device.uid, device.name)}
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

      {device.status === "accepted" && (
        <SshidCard uid={device.uid} name={device.name} className="mb-6" />
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
        <IdentityCard
          uid={device.uid}
          mac={device.identity?.mac ?? ""}
          remoteAddr={device.remote_addr ?? ""}
          registeredVia={
            enrollment && device.provisioning_key_id ? (
              <Link
                to={provisioningKeyActivityPath(device.provisioning_key_id)}
                className="text-sm font-medium text-text-primary hover:text-primary hover:underline"
              >
                {enrollmentSourceName(enrollment)}
              </Link>
            ) : (
              <span className="text-sm text-text-muted">—</span>
            )
          }
          pairedBy={
            <PairedBy
              tenantId={tenantId}
              uid={device.uid}
              name={device.name}
              ownerId={device.owner_id}
            />
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
              value={device.info?.pretty_name}
            />
            <InfoItem
              label="Architecture"
              value={device.info?.arch}
              mono
            />
            <InfoItem label="Platform">
              {device.info?.platform && (
                <PlatformBadge platform={device.info.platform} />
              )}
            </InfoItem>
            <InfoItem
              label="Agent Version"
              value={device.info?.version}
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

      {actionsController.action && (
        <ActionDialog
          key={actionsController.actionKey}
          action={actionsController.action}
          onClose={actionsController.close}
          onSuccess={actionsController.handleSuccess}
          entityType="device"
          runAction={runDeviceAction}
        />
      )}
    </div>
  );
}
