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

interface Props {
  open: boolean;
  onClose: () => void;
  /** Called when the user chooses to sync; opens the sync flow. */
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
    <BaseDialog open={open} onClose={close} size="sm" aria-labelledby={titleId}>
      <div className="p-6">
        <div className="flex flex-col items-center text-center mb-5">
          <div className="relative w-14 h-14 rounded-2xl bg-primary/10 flex items-center justify-center mb-4">
            <div className="absolute inset-0 rounded-2xl bg-primary/20 blur-xl -z-10" />
            <ServerStackIcon className="w-7 h-7 text-primary" />
          </div>
          <h2
            id={titleId}
            className="text-base font-semibold text-text-primary"
          >
            Take your vault anywhere
          </h2>
          <p className="text-sm text-text-secondary mt-1.5 max-w-xs">
            This vault lives in this browser only. Sync it to the ShellHub
            server and it follows you.
          </p>
        </div>

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
    </BaseDialog>
  );
}
