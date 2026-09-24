import { useState } from "react";
import { useUpdateProvisioningKey } from "@/hooks/useProvisioningKeyMutations";
import { type ProvisioningKey } from "@/client";

/**
 * Enables and disables a provisioning key, holding the failure so the row can show it. Disabling is
 * reversible, which is what distinguishes it from revoking.
 */
export function useToggleProvisioningKey() {
  const updateKey = useUpdateProvisioningKey();
  const [error, setError] = useState<string | null>(null);

  const toggle = async (key: ProvisioningKey) => {
    setError(null);
    try {
      await updateKey.mutateAsync({
        path: { key: key.name },
        body: { disabled: !key.disabled },
      });
    } catch {
      setError(
        `Failed to ${key.disabled ? "enable" : "disable"} Provisioning Key.`,
      );
    }
  };

  return { toggle, error, isToggling: updateKey.isPending };
}
