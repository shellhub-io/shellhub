import { Link } from "react-router-dom";
import {
  ChevronDoubleRightIcon,
  CpuChipIcon,
  TagIcon,
} from "@heroicons/react/24/outline";
import { Badge, Button, StatusDot } from "@shellhub/design-system/primitives";
import type { Device } from "@/client";
import Modal from "@/components/common/Modal";
import InfoItem from "@/components/common/InfoItem";
import PlatformBadge from "@/components/common/PlatformBadge";
import RestrictedAction from "@/components/common/RestrictedAction";
import SshidCard from "@/components/common/SshidCard";
import { useTerminalStore } from "@/stores/terminalStore";
import { formatRelative } from "@/utils/date";
import { normalizeDeviceTags } from "@/utils/deviceTags";

interface DeviceBriefDialogProps {
  device: Device;
  onClose: () => void;
}

/**
 * A short summary of a device picked on the map, from the device as the area list read it, with
 * a link to its details. Connect resumes the device's terminal session when one is open, and
 * otherwise opens the connect dialog. onClose runs when this dialog closes, Connect included.
 */
export default function DeviceBriefDialog({
  device,
  onClose,
}: DeviceBriefDialogProps) {
  const connectTo = useTerminalStore((s) => s.connect);
  const { tags } = normalizeDeviceTags(device);

  const connect = () => {
    connectTo(device.uid, device.name);
    onClose();
  };

  return (
    <Modal
      open
      onClose={onClose}
      icon={<CpuChipIcon />}
      title={device.name}
      description={
        <span className="flex flex-wrap items-center gap-x-3 gap-y-0.5">
          <span className="inline-flex items-center gap-1.5">
            <StatusDot online={device.online} size="sm" />
            {device.online ? "Online" : "Offline"}
          </span>
          {device.info?.pretty_name && <span>{device.info.pretty_name}</span>}
        </span>
      }
      footerStart={
        <Link
          to={`/devices/${device.uid}`}
          className="transition-colors hover:text-text-primary"
        >
          Open device details &rarr;
        </Link>
      }
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>
            Close
          </Button>
          <RestrictedAction action="device:connect">
            <Button
              variant="success"
              onClick={connect}
              disabled={!device.online}
              icon={
                <ChevronDoubleRightIcon className="w-4 h-4" strokeWidth={2} />
              }
            >
              Connect
            </Button>
          </RestrictedAction>
        </>
      }
    >
      <SshidCard uid={device.uid} name={device.name} />

      <dl className="mt-5 grid grid-cols-2 gap-4">
        <InfoItem label="Platform">
          {device.info?.platform ? (
            <span className="inline-flex items-center gap-2">
              <PlatformBadge platform={device.info.platform} />
              <span className="font-mono text-xs text-text-muted">
                {device.info.arch}
              </span>
            </span>
          ) : (
            <span className="text-sm text-text-muted">—</span>
          )}
        </InfoItem>
        <InfoItem
          label="Agent Version"
          value={device.info?.version ?? ""}
          mono
        />
        <InfoItem
          label="Last Seen"
          value={formatRelative(device.last_seen)}
        />
        <InfoItem label="Tags">
          {tags.length > 0 ? (
            <span className="flex flex-wrap gap-1">
              {tags.map((tag) => (
                <Badge key={tag} color="primary">
                  <TagIcon className="w-2.5 h-2.5" strokeWidth={2} />
                  {tag}
                </Badge>
              ))}
            </span>
          ) : (
            <span className="text-sm text-text-muted">—</span>
          )}
        </InfoItem>
      </dl>
    </Modal>
  );
}
