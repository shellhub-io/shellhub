import { useState } from "react";
import {
  NoSymbolIcon,
} from "@heroicons/react/24/outline";
import { useUpdateProvisioningKey } from "@/hooks/useProvisioningKeyMutations";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { type ProvisioningKey } from "@/client";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import InputField from "@/components/common/fields/InputField";
import ObjectName from "@/components/common/ObjectName";

/**
 * Confirms revoking a provisioning key. Revoking is final — unlike disabling, it cannot be undone —
 * which is what the dialog has to make clear.
 */
export default function RevokeProvisioningKeyDialog({
  provisioningKey,
  onRevoked,
}: {
  provisioningKey: ProvisioningKey | null;
  onRevoked: () => void;
}) {
  const updateKey = useUpdateProvisioningKey();
  const [confirmText, setConfirmText] = useState("");
  const [error, setError] = useState<string | null>(null);

  const open = !!provisioningKey;
  const name = provisioningKey?.name ?? "";

  useResetOnOpen(open, () => {
    setConfirmText("");
    setError(null);
  });

  const handleConfirm = async () => {
    if (!provisioningKey) return;
    setError(null);
    try {
      await updateKey.mutateAsync({
        path: { key: provisioningKey.name },
        body: { revoked: true },
      });
      onRevoked();
    } catch {
      setError("Failed to revoke Provisioning Key.");
    }
  };

  return (
    <ConfirmDialog
      open={open}
      onClose={onRevoked}
      onConfirm={handleConfirm}
      icon={<NoSymbolIcon />}
      title="Revoke provisioning key"
      description={
        <>
          Devices and pipelines can no longer register with{" "}
          <ObjectName>{name}</ObjectName>; the ones it already registered keep
          working. This can't be undone. Type <ObjectName>{name}</ObjectName> to
          confirm.
        </>
      }
      confirmLabel="Revoke key"
      confirmDisabled={confirmText !== name}
      errorMessage={error}
    >
      <InputField
        id="revoke-provisioning-key-confirm"
        label="Type the key's name to confirm"
        hideLabel
        value={confirmText}
        onChange={setConfirmText}
        autoComplete="off"
      />
    </ConfirmDialog>
  );
}
