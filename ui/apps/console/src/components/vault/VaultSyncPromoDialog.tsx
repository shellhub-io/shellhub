import { useState, useId } from "react";
import { Button } from "@shellhub/design-system/primitives";
import {
  ServerStackIcon,
  GlobeAltIcon,
  ShieldCheckIcon,
  ArrowPathIcon,
} from "@heroicons/react/24/outline";
import { useAuthStore } from "@/stores/authStore";
import { dismissVaultSyncPromo } from "@/utils/vault-backend-factory";
import BaseDialog from "@/components/common/BaseDialog";
import CheckboxField from "@/components/common/fields/CheckboxField";
import DialogHeader from "@/components/common/DialogHeader";

interface Props {
  open: boolean;
  onClose: () => void;
  onSync: () => void;
}

const BENEFITS = [
  {
    icon: GlobeAltIcon,
    text: "Unlock your keys from any machine you sign in to",
  },
  {
    icon: ShieldCheckIcon,
    text: "End-to-end encrypted. The server never sees your keys",
  },
  {
    icon: ArrowPathIcon,
    text: "Survives clearing this browser's data",
  },
];

/**
 * Offers server-side sync to someone whose vault is local only. Dismissable for good, since it
 * is an offer and not a requirement.
 */
export default function VaultSyncPromoDialog({ open, onClose, onSync }: Props) {
  const instanceId = useId();
  const titleId = `vault-sync-promo-title-${instanceId}`;
  const user = useAuthStore((s) => s.user);
  const tenant = useAuthStore((s) => s.tenant);
  const [dontShowAgain, setDontShowAgain] = useState(false);

  const close = () => {
    if (dontShowAgain) {
      dismissVaultSyncPromo(user && tenant ? { user, tenant } : undefined);
    }
    onClose();
  };

  return (
    <BaseDialog
open={open}
onClose={close}
size="sm"
aria-labelledby={titleId}
      aria-describedby={`${titleId}-description`}
    >
      <div>
        <DialogHeader
          layout="center"
          icon={<ServerStackIcon />}
          title="Take your vault anywhere"
          description="This vault lives in this browser only. Sync it to the ShellHub server and it follows you."
          titleId={titleId}
          descriptionId={`${titleId}-description`}
          onClose={close}
        />
        <div className="px-8 pb-6">

        <ul className="space-y-2.5 mb-5">
          {BENEFITS.map(({ icon: Icon, text }) => (
            <li key={text} className="flex items-center gap-3">
              <Icon className="w-4 h-4 text-primary shrink-0" strokeWidth={2} />
              <span className="text-xs text-text-secondary">{text}</span>
            </li>
          ))}
        </ul>

        <div className="flex flex-col gap-3">
          <Button
            fullWidth
            onClick={() => {
              onClose();
              onSync();
            }}
          >
            Sync vault
          </Button>
          <Button variant="ghost" fullWidth onClick={close}>
            Keep it on this device
          </Button>
        </div>

        <div className="flex justify-center mt-4">
          <CheckboxField
            id={`${instanceId}-dont-show`}
            label="Don't show this again"
            checked={dontShowAgain}
            onChange={setDontShowAgain}
          />
        </div>
        </div>
      </div>
    </BaseDialog>
  );
}
