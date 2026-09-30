import { useState } from "react";
import { isSdkError } from "@/api/errors";
import { useAuthStore } from "@/stores/authStore";
import { useNamespace } from "@/hooks/useNamespaces";
import { useDevices } from "@/hooks/useDevices";
import {
  useAcceptDevicePairing,
  useResolveDeviceCode,
} from "@/hooks/useDeviceCode";
import { useAcceptDevice } from "@/hooks/useDeviceMutations";
import { useHasPermission } from "@/hooks/useHasPermission";
import { isSubscriptionBlocked } from "@/utils/billing";
import { getAcceptErrorMessage } from "@/utils/acceptErrors";

const LINK_POLL_MS = 3000;

const UNKNOWN_CODE =
  "That code is invalid or has expired. Run shellhub-agent login on the device to get a new one.";

const LOOKUP_FAILED =
  "Couldn't look that code up. Check your connection and try again.";

const OTHER_NAMESPACE =
  "That device is waiting in another namespace. Accept it from that namespace's device list.";

/**
 * A device the first run has paired: enough to name it and connect to it.
 */
export interface PairedDevice {
  uid: string;
  name: string;
}

/**
 * Where getting the first device into the current namespace stands: waiting for a code, showing
 * what a code resolved to, or done with a device to open a shell on.
 */
export type FirstDeviceStage = "install" | "pair" | "shell";

/**
 * Drives getting the first device into the current namespace: the code the user entered, what it
 * resolved to, pairing it, and the device that came of it. A device accepted from the printed
 * link, rather than through the code, is picked up by polling the namespace's accepted devices.
 */
export function useFirstDevice() {
  const tenant = useAuthStore((s) => s.tenant);
  const { namespace } = useNamespace(tenant ?? "");
  const canSubscribe = useHasPermission("billing:subscribe");

  const [code, setCode] = useState("");
  const [paired, setPaired] = useState<PairedDevice | null>(null);
  const [acceptError, setAcceptError] = useState("");

  const {
    device: preview,
    isFetching: resolving,
    isError,
    error,
    refetch: retryLookup,
  } = useResolveDeviceCode(code);
  const acceptPairing = useAcceptDevicePairing();
  const acceptDevice = useAcceptDevice();
  const { devices: accepted } = useDevices({
    status: "accepted",
    perPage: 1,
    refetchInterval: (page) =>
      paired || page?.data.length ? false : LINK_POLL_MS,
  });

  const linked = accepted[0]
    ? { uid: accepted[0].uid, name: accepted[0].name }
    : null;
  const codeError = (() => {
    if (!code || resolving) return "";
    if (isError)
      return isSdkError(error) && error.status === 404
        ? UNKNOWN_CODE
        : LOOKUP_FAILED;
    if (preview?.kind === "device" && preview.tenant_id !== tenant)
      return OTHER_NAMESPACE;
    return "";
  })();
  const pairable = preview && !codeError ? preview : null;
  const device = paired ?? linked;
  const stage: FirstDeviceStage = device
    ? "shell"
    : pairable
      ? "pair"
      : "install";

  const submitCode = (next: string) => {
    if (next === code && isError) void retryLookup();
    else setCode(next);
  };

  const backToCode = () => {
    setCode("");
    setAcceptError("");
  };

  const pair = async () => {
    if (!pairable) return;
    setAcceptError("");
    try {
      if (pairable.kind === "pairing") {
        const data = await acceptPairing.mutateAsync({
          path: { code },
          body: { tenant_id: tenant ?? "" },
        });
        setPaired({ uid: data.uid ?? "", name: pairable.name ?? "" });
      } else if (pairable.uid) {
        await acceptDevice.mutateAsync({ path: { uid: pairable.uid } });
        setPaired({ uid: pairable.uid, name: pairable.name ?? "" });
      }
    } catch (err) {
      setAcceptError(
        getAcceptErrorMessage(
          err,
          isSubscriptionBlocked(namespace?.billing),
          canSubscribe,
        ),
      );
    }
  };

  return {
    namespace: namespace?.name ?? "",
    stage,
    code,
    submitCode,
    resolving,
    codeError,
    pairable,
    pair: () => void pair(),
    pairing: acceptPairing.isPending || acceptDevice.isPending,
    acceptError,
    backToCode,
    paired: paired !== null,
    device,
  };
}
