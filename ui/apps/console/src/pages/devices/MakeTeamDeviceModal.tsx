import { useState } from "react";
import {
  ArrowRightIcon,
  CheckCircleIcon,
  UserGroupIcon,
} from "@heroicons/react/24/outline";
import { Button, Callout, IconBadge } from "@shellhub/design-system/primitives";
import { apiErrorMessage } from "@/api/errors";
import { useMakeTeamDevice } from "@/hooks/useDeviceMutations";
import Modal from "@/components/common/Modal";
import ObjectName from "@/components/common/ObjectName";
import UserBadge from "@/components/common/UserBadge";
import { LABEL, LABEL_BASE } from "@/utils/styles";

const UNCHANGED = [
  "The device stays accepted and connected. Its agent is not told.",
  "Registered via still shows that it arrived by pairing.",
  "Sessions, tags and custom fields are kept.",
];

/**
 * Confirms making a paired device a team device: it stops leaving with the member who paired it
 * and stays in the namespace whoever leaves. Nothing else about the device changes.
 */
export default function MakeTeamDeviceModal({
  uid,
  name,
  ownerName,
  ownerEmail,
  onClose,
}: {
  uid: string;
  name: string;
  ownerName?: string;
  ownerEmail: string;
  onClose: () => void;
}) {
  const makeTeamDevice = useMakeTeamDevice();
  const [error, setError] = useState<string | null>(null);

  const confirm = async () => {
    setError(null);
    try {
      await makeTeamDevice.mutateAsync({ path: { uid } });
      onClose();
    } catch (err) {
      setError(apiErrorMessage(err));
    }
  };

  return (
    <Modal
      open
      onClose={onClose}
      icon={<UserGroupIcon />}
      title="Make team device"
      description={
        <>
          <ObjectName>{name}</ObjectName> stops belonging to one member and
          stays in this namespace whoever leaves it.
        </>
      }
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button
            variant="primary"
            icon={<UserGroupIcon className="w-4 h-4" />}
            loading={makeTeamDevice.isPending}
            onClick={() => void confirm()}
          >
            Make team device
          </Button>
        </>
      }
    >
      <div className="space-y-6">
        <div className="grid grid-cols-1 sm:grid-cols-[1fr_auto_1fr] items-stretch gap-3">
          <div className="rounded-lg border border-border bg-card p-4">
            <p className={LABEL}>Now</p>
            <UserBadge name={ownerName} email={ownerEmail} />
            <p className="mt-3 text-2xs text-text-muted">
              Removed when this member leaves or can no longer accept devices.
            </p>
          </div>
          <ArrowRightIcon className="hidden sm:block self-center w-4 h-4 text-text-muted" />
          <div className="rounded-lg border border-primary/30 bg-primary/[0.05] p-4">
            <p className={LABEL}>After</p>
            <span className="inline-flex items-center gap-2.5">
              <IconBadge size="sm" className="[&>svg]:w-4 [&>svg]:h-4">
                <UserGroupIcon />
              </IconBadge>
              <span className="text-sm font-medium text-text-primary">
                Team device
              </span>
            </span>
            <p className="mt-3 text-2xs text-text-muted">
              Stays in the namespace whoever leaves it.
            </p>
          </div>
        </div>

        <section aria-label="What stays the same" className="space-y-2.5">
          <h3 className={LABEL_BASE}>What stays the same</h3>
          <ul className="space-y-2">
            {UNCHANGED.map((text) => (
              <li
                key={text}
                className="flex items-start gap-2 text-xs text-text-secondary"
              >
                <CheckCircleIcon className="w-4 h-4 shrink-0 mt-px text-accent-green" />
                {text}
              </li>
            ))}
          </ul>
        </section>

        <Callout variant="warning">
          This cannot be undone. A team device cannot be given an owner again.
        </Callout>

        {error && <Callout variant="error">{error}</Callout>}
      </div>
    </Modal>
  );
}
