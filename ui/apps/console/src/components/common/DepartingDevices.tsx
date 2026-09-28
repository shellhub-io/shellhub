import { ArrowPathIcon, CpuChipIcon } from "@heroicons/react/24/outline";
import {
  Badge,
  Callout,
  IconBadge,
  Spinner,
} from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { MAX_KEPT_DEVICES, useMemberDevices } from "@/hooks/useMemberDevices";
import CheckboxField from "@/components/common/fields/CheckboxField";
import { LABEL_BASE } from "@/utils/styles";

const LISTED_DEVICES = 5;

function devicesLabel(count: number) {
  return count === 1 ? "1 device" : `${count} devices`;
}

interface DepartingDevicesProps {
  memberId: string;
  self?: boolean;
  keep?: string[];
  onKeepChange?: (keep: string[]) => void;
}

/**
 * The devices a member paired, which leave the namespace with them. When onKeepChange is given,
 * up to MAX_KEPT_DEVICES are listed and can be kept as team devices, and any beyond that are said
 * to leave; otherwise the first few are named and the rest counted. It shows a loading and an
 * error state, which the dialog around it has to hold its confirmation for, and renders nothing
 * for a member who paired none.
 */
export default function DepartingDevices({
  memberId,
  self = false,
  keep = [],
  onKeepChange,
}: DepartingDevicesProps) {
  const { devices, totalCount, isLoading, isError } = useMemberDevices(memberId);
  const whose = self ? "you" : "this member";

  if (isLoading) {
    return (
      <p className="flex items-center gap-2 text-xs text-text-muted">
        <Spinner size="sm" />
        Loading the devices {whose} paired...
      </p>
    );
  }

  if (isError) {
    return (
      <Callout variant="error">
        Couldn&apos;t load the devices {whose} paired, so they can&apos;t be
        shown before they are removed. Try again.
      </Callout>
    );
  }

  if (totalCount === 0) return null;

  const leaving = totalCount - keep.length;
  const listed = onKeepChange ? devices : devices.slice(0, LISTED_DEVICES);
  const unlisted = totalCount - listed.length;

  const toggle = (uid: string, kept: boolean) =>
    onKeepChange?.(kept ? [...keep, uid] : keep.filter((k) => k !== uid));

  return (
    <section aria-label="Paired devices" className="space-y-2.5">
      <div className="flex items-center justify-between gap-3">
        <h3 className={cn(LABEL_BASE, "flex items-center gap-2")}>
          Paired devices
          <Badge shape="pill" color="yellow">
            {totalCount}
          </Badge>
        </h3>
        {onKeepChange && (
          <span className={LABEL_BASE}>Keep as team device</span>
        )}
      </div>

      <ul className="rounded-lg border border-border divide-y divide-border overflow-hidden">
        {listed.map((device) => {
          const kept = keep.includes(device.uid);

          return (
            <li
              key={device.uid}
              className={cn(
                "flex items-center gap-3 px-3.5 py-2.5 transition-colors",
                kept ? "bg-primary/[0.04]" : "bg-card",
              )}
            >
              <IconBadge
                size="sm"
                color={kept ? "primary" : "yellow"}
                className="[&>svg]:w-4 [&>svg]:h-4"
              >
                <CpuChipIcon />
              </IconBadge>
              <div className="min-w-0 flex-1">
                <p className="text-sm font-mono text-text-primary truncate">
                  {device.name}
                </p>
                <p className="text-2xs text-text-muted truncate">
                  {device.info?.pretty_name ?? "Unknown system"}
                </p>
              </div>
              <Badge color={kept ? "primary" : "yellow"}>
                {kept ? "Stays" : "Leaves"}
              </Badge>
              {onKeepChange && (
                <CheckboxField
                  id={`keep-${device.uid}`}
                  label="Keep as team device"
                  hideLabel
                  checked={kept}
                  onChange={(next) => toggle(device.uid, next)}
                  aria-label={`Keep ${device.name} as a team device`}
                />
              )}
            </li>
          );
        })}
      </ul>

      {unlisted > 0 && (
        <p className="text-2xs text-text-muted">
          {onKeepChange
            ? `and ${unlisted} more, which leave: only the first ${MAX_KEPT_DEVICES} can be kept here`
            : `and ${unlisted} more`}
        </p>
      )}

      <p className="flex items-start gap-2 text-xs text-text-secondary">
        <ArrowPathIcon className="w-4 h-4 shrink-0 text-accent-yellow mt-px" />
        <span>
          {leaving === 0
            ? "Every device this member paired stays as a team device."
            : self
              ? `${devicesLabel(leaving)} you paired will be removed with you and go back to pairing.`
              : `${devicesLabel(leaving)} this member paired will be removed and go back to pairing.`}
        </span>
      </p>
    </section>
  );
}
