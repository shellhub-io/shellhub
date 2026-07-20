import { useState } from "react";
import { resolveDeviceLoginCode, acceptDevicePairing } from "@/client";
import { isSdkError } from "@/api/errors";
import { useAuthStore } from "@/stores/authStore";
import { useAcceptDevice } from "@/hooks/useDeviceMutations";
import { getAcceptDeviceErrorMessage } from "@/utils/deviceErrors";

/**
 * Resolves a pairing/login code and accepts the device in one shot, into the
 * current namespace. Used by the onboarding wizard's code-entry step, which
 * skips the preview/confirm screen: the user just installed the device, so they
 * type the code and go. Returns the accepted device, or null with `error` set.
 */
export function useAcceptDeviceByCode() {
  const authTenant = useAuthStore((s) => s.tenant);
  const acceptDevice = useAcceptDevice();
  const [isPending, setIsPending] = useState(false);
  const [error, setError] = useState("");

  const submit = async (
    code: string,
  ): Promise<{ uid: string; name: string } | null> => {
    setError("");
    setIsPending(true);
    try {
      const { data } = await resolveDeviceLoginCode({
        path: { code },
        throwOnError: true,
      });

      // Tenant-less pairing code (what the codeless install prints): the device
      // doesn't exist yet, so accept it into the current namespace.
      if (data.kind === "pairing") {
        const { data: accepted } = await acceptDevicePairing({
          path: { code },
          body: { tenant_id: authTenant ?? "" },
          throwOnError: true,
        });
        return { uid: accepted.uid ?? "", name: data.name ?? "" };
      }

      // Login code for a device that already registered as pending.
      if (data.uid) {
        await acceptDevice.mutateAsync({ path: { uid: data.uid } });
        return { uid: data.uid, name: data.name ?? "" };
      }

      setError(
        "That code can't be used here. Open the link the agent printed.",
      );
      return null;
    } catch (err) {
      // The API collapses unknown, wrong, and expired codes into 404 on
      // purpose (the code is the secret), so this one message covers all three.
      setError(
        isSdkError(err) && err.status === 404
          ? "That code is invalid or has expired. Double-check it and try again."
          : getAcceptDeviceErrorMessage(err),
      );
      return null;
    } finally {
      setIsPending(false);
    }
  };

  return { submit, isPending, error, clearError: () => setError("") };
}
