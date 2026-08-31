import { useState } from "react";
import { useUpdateInstallKey } from "@/hooks/useInstallKeyMutations";
import { type InstallKey } from "@/client";

/**
 * Enables and disables an install key, holding the failure so the row can show it. Disabling is
 * reversible, which is what distinguishes it from revoking.
 */
export function useToggleInstallKey() {
  const updateKey = useUpdateInstallKey();
  const [error, setError] = useState<string | null>(null);

  const toggle = async (key: InstallKey) => {
    setError(null);
    try {
      await updateKey.mutateAsync({
        path: { key: key.name },
        body: { disabled: !key.disabled },
      });
    } catch {
      setError(
        `Failed to ${key.disabled ? "enable" : "disable"} Install Key.`,
      );
    }
  };

  return { toggle, error, isToggling: updateKey.isPending };
}
