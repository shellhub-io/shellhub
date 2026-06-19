import { LockClosedIcon } from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";

interface Props {
  onUnlock: () => void;
}

export default function VaultLockedBanner({ onUnlock }: Props) {
  return (
    <div className="flex items-center justify-between bg-accent-yellow/[0.06] border border-accent-yellow/20 rounded-lg px-4 py-3">
      <div className="flex items-center gap-3">
        <LockClosedIcon className="w-5 h-5 text-accent-yellow shrink-0" />
        <p className="text-sm text-text-secondary">
          Your vault is locked. Unlock it to view and manage your SSH keys.
        </p>
      </div>
      <Button
        size="sm"
        onClick={onUnlock}
        aria-label="Unlock vault to access SSH keys"
        className="shrink-0"
      >
        Unlock
      </Button>
    </div>
  );
}
