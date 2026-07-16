import { type ReactNode } from "react";
import { InformationCircleIcon } from "@heroicons/react/24/outline";
import { Card } from "@shellhub/design-system/primitives";
import InfoItem from "@/components/common/InfoItem";

interface IdentityCardProps {
  uid: string;
  mac: string;
  remoteAddr: string;
  // Optional "Registered via" row: which install key the device registered with. Shown only where the
  // caller provides it (the device details page); other callers omit it.
  registeredVia?: ReactNode;
}

export default function IdentityCard({
  uid,
  mac,
  remoteAddr,
  registeredVia,
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
      </dl>
    </Card>
  );
}
