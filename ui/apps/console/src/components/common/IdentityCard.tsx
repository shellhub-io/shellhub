import { type ReactNode } from "react";
import { InformationCircleIcon } from "@heroicons/react/24/outline";
import { Card } from "@shellhub/design-system/primitives";
import InfoItem from "@/components/common/InfoItem";
import type { DeviceReplacement } from "@/client";
import { formatDateFull } from "@/utils/date";

interface IdentityCardProps {
  uid: string;
  mac: string;
  remoteAddr: string;
  registeredVia?: ReactNode;
  pairedBy?: ReactNode;
  replaced?: DeviceReplacement;
}

/**
 * The identifying facts about a device — UID, MAC, address — as a copyable block, since these
 * are values people paste into a terminal rather than read. replaced, when the device replaced an
 * offline one with the same identity on accept, names that device in full and says when.
 */
export default function IdentityCard({
  uid,
  mac,
  remoteAddr,
  registeredVia,
  pairedBy,
  replaced,
}: IdentityCardProps) {
  return (
    <Card className="p-5 space-y-4">
      <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2">
        <InformationCircleIcon className="w-4 h-4 text-primary" />
        Identity
      </h3>
      <dl className="space-y-3">
        <InfoItem label="UID" value={uid} mono copyable truncate={8} />
        <InfoItem label="MAC Address" value={mac} mono copyable />
        <InfoItem label="Remote Address" value={remoteAddr} mono />
        {registeredVia && (
          <InfoItem label="Registered via">{registeredVia}</InfoItem>
        )}
        {pairedBy && <InfoItem label="Paired by">{pairedBy}</InfoItem>}
        {replaced && (
          <InfoItem label="Replaced">
            <span className="text-sm text-text-primary">
              <span className="font-mono">{replaced.name}</span>{" "}
              <span className="font-mono text-text-muted break-all">
                ({replaced.uid})
              </span>{" "}
              on {formatDateFull(replaced.at)}
            </span>
          </InfoItem>
        )}
      </dl>
    </Card>
  );
}
