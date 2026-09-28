import { useState } from "react";
import {
  KeyIcon,
  NoSymbolIcon,
  UserMinusIcon,
} from "@heroicons/react/24/outline";
import { Button, Callout } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { apiErrorMessage } from "@/api/errors";
import type { MemberView } from "@/client";
import { useHasPermission } from "@/hooks/useHasPermission";
import { useMemberDevices } from "@/hooks/useMemberDevices";
import { useRemoveMember } from "@/hooks/useMemberMutations";
import DepartingDevices from "@/components/common/DepartingDevices";
import Modal from "@/components/common/Modal";
import ObjectName from "@/components/common/ObjectName";
import UserBadge from "@/components/common/UserBadge";
import { LABEL_BASE } from "@/utils/styles";
import { RoleBadge } from "./constants";

const CONSEQUENCES = [
  {
    icon: NoSymbolIcon,
    text: "Loses access to this namespace, its devices and its sessions.",
  },
  {
    icon: KeyIcon,
    text: "The API keys they created in this namespace are revoked.",
  },
];

/**
 * Removes a member from the namespace, showing what goes with them: the devices they paired,
 * each of which an admin may keep as a team device instead, and the API keys they created.
 */
export default function RemoveMemberModal({
  tenantId,
  member,
  onClose,
}: {
  tenantId: string;
  member: MemberView;
  onClose: () => void;
}) {
  const removeMember = useRemoveMember();
  const devices = useMemberDevices(member.id ?? "");
  const canKeepDevices = useHasPermission("provisioningKey:create");
  const [keepDevices, setKeepDevices] = useState<string[]>([]);
  const [error, setError] = useState<string | null>(null);

  const remove = async () => {
    if (!member.id) return;
    setError(null);
    try {
      await removeMember.mutateAsync({
        path: { tenant: tenantId, uid: member.id },
        query:
          keepDevices.length > 0 ? { keep_devices: keepDevices } : undefined,
      });
      onClose();
    } catch (err) {
      setError(apiErrorMessage(err));
    }
  };

  return (
    <Modal
      open
      onClose={onClose}
      icon={<UserMinusIcon />}
      iconColor="red"
      title="Remove member"
      description={
        <>
          <ObjectName>{member.email}</ObjectName> leaves this namespace.
          Someone has to invite them again to bring them back.
        </>
      }
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button
            variant="destructive"
            icon={<UserMinusIcon className="w-4 h-4" />}
            loading={removeMember.isPending}
            disabled={devices.isLoading || devices.isError}
            onClick={() => void remove()}
          >
            Remove member
          </Button>
        </>
      }
    >
      <div className="space-y-6">
        <div className="flex items-center justify-between gap-3 rounded-lg border border-border bg-card px-4 py-3">
          <UserBadge name={member.name} email={member.email} />
          <RoleBadge role={member.role ?? "observer"} />
        </div>

        <DepartingDevices
          memberId={member.id ?? ""}
          keep={keepDevices}
          onKeepChange={canKeepDevices ? setKeepDevices : undefined}
        />

        <section aria-label="What else happens" className="space-y-2.5">
          <h3 className={LABEL_BASE}>What else happens</h3>
          <ul className="space-y-2">
            {CONSEQUENCES.map(({ icon: Icon, text }) => (
              <li
                key={text}
                className="flex items-start gap-2 text-xs text-text-secondary"
              >
                <Icon className={cn("w-4 h-4 shrink-0 mt-px text-accent-red")} />
                {text}
              </li>
            ))}
          </ul>
        </section>

        {error && <Callout variant="error">{error}</Callout>}
      </div>
    </Modal>
  );
}
