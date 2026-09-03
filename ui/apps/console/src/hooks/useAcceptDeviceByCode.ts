import { useState } from "react";
import {
  resolveDeviceLoginCode,
  acceptDevicePairing,
  useAcceptDevice,
} from "@/client/api";
import { isSdkError } from "@/api/errors";
import { useAuthStore } from "@/stores/authStore";
import { useHasPermission } from "@/hooks/useHasPermission";
import { useNamespace } from "@/hooks/useNamespaces";
import { isSubscriptionBlocked } from "@/utils/billing";
import { getAcceptErrorMessage } from "@/utils/acceptErrors";

/** Resolves a pairing/login code and accepts the device into the current namespace. */
export function useAcceptDeviceByCode() {
  const authTenant = useAuthStore((s) => s.tenant) ?? "";
  const { namespace } = useNamespace(authTenant);
  const hasSubscription = isSubscriptionBlocked(namespace?.billing);
  const canSubscribe = useHasPermission("billing:subscribe");
  const acceptDevice = useAcceptDevice();
  const [isPending, setIsPending] = useState(false);
  const [error, setError] = useState("");

  const submit = async (
    code: string,
  ): Promise<{ uid: string; name: string } | null> => {
    setError("");
    setIsPending(true);
    try {
      const resolved = await resolveDeviceLoginCode(code);

      if (resolved.kind === "pairing") {
        const accepted = await acceptDevicePairing(code, {
          tenant_id: authTenant,
        });
        return { uid: accepted.uid ?? "", name: resolved.name ?? "" };
      }

      if (resolved.uid) {
        await acceptDevice.mutateAsync({ uid: resolved.uid });
        return { uid: resolved.uid, name: resolved.name ?? "" };
      }

      setError(
        "That code can't be used here. Open the link the agent printed.",
      );
      return null;
    } catch (err) {
      setError(
        isSdkError(err) && err.status === 404
          ? "That code is invalid or has expired. Double-check it and try again."
          : getAcceptErrorMessage(err, hasSubscription, canSubscribe),
      );
      return null;
    } finally {
      setIsPending(false);
    }
  };

  return { submit, isPending, error, clearError: () => setError("") };
}
