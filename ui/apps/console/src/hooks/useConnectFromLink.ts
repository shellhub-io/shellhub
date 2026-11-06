import { useEffect } from "react";
import { useSearchParams } from "react-router-dom";
import { useHasPermission } from "@/hooks/useHasPermission";
import { useTerminalStore } from "@/stores/terminalStore";

interface Connectable {
  uid: string;
  name: string;
  online: boolean;
}

/**
 * Acts once when the page's URL carries connect=true: brings back the device's open terminal when
 * there is one, and otherwise opens the connect dialog. Then drops the parameter, so a refetch, a
 * rename or a reload does not ask again. Does nothing while the device is missing or offline, or
 * when the role lacks device:connect; the parameter stays until it can act on it.
 */
export function useConnectFromLink(device: Connectable | null | undefined) {
  const [searchParams, setSearchParams] = useSearchParams();
  const canConnect = useHasPermission("device:connect");
  const connect = useTerminalStore((s) => s.connect);

  const target =
    canConnect && searchParams.get("connect") === "true" && device?.online
      ? device
      : undefined;
  const uid = target?.uid;
  const name = target?.name;

  useEffect(() => {
    if (!uid || name === undefined) return;
    connect(uid, name);
    setSearchParams(
      (params) => {
        params.delete("connect");
        return params;
      },
      { replace: true },
    );
  }, [uid, name, connect, setSearchParams]);
}
