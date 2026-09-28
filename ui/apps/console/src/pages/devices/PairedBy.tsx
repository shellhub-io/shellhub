import { useState } from "react";
import { useHasPermission } from "@/hooks/useHasPermission";
import { useNamespaceMembers } from "@/hooks/useNamespaces";
import MakeTeamDeviceModal from "./MakeTeamDeviceModal";

/**
 * Who paired a device, whose departure would remove it, with the action that makes it a team
 * device for those allowed to create provisioning keys. A team device shows a dash.
 */
export default function PairedBy({
  tenantId,
  uid,
  name,
  ownerId,
}: {
  tenantId: string;
  uid: string;
  name: string;
  ownerId?: string;
}) {
  const { members } = useNamespaceMembers(ownerId ? tenantId : "");
  const canMakeTeamDevice = useHasPermission("provisioningKey:create");
  const [confirming, setConfirming] = useState(false);

  if (!ownerId) {
    return <span className="text-sm text-text-muted">—</span>;
  }

  const owner = members.find((m) => m.id === ownerId);
  const ownerEmail = owner?.email ?? ownerId;

  return (
    <span className="flex flex-wrap items-center gap-x-3 gap-y-1">
      <span className="text-sm font-medium text-text-primary truncate">
        {ownerEmail}
      </span>
      {canMakeTeamDevice && (
        <button
          type="button"
          onClick={() => setConfirming(true)}
          className="text-xs text-primary hover:underline"
        >
          Make team device
        </button>
      )}
      {confirming && (
        <MakeTeamDeviceModal
          uid={uid}
          name={name}
          ownerName={owner?.name}
          ownerEmail={ownerEmail}
          onClose={() => setConfirming(false)}
        />
      )}
    </span>
  );
}
