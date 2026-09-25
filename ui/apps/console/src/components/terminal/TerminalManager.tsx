import { useEffect, useRef, useState } from "react";
import { useLocation } from "react-router-dom";
import { cn } from "@shellhub/design-system/cn";
import { useTerminalStore } from "@/stores/terminalStore";
import { useNamespaces } from "@/hooks/useNamespaces";
import { useWorkspaceTabs } from "@/hooks/useWorkspaceTabs";
import { useAuthStore } from "@/stores/authStore";
import ConnectDrawer from "../ConnectDrawer";
import { buildSshid } from "@/utils/sshid";
import TerminalInstance from "./TerminalInstance";
import RecordingSnackbar from "./RecordingSnackbar";

/**
 * Holds every open terminal window, stacked over the page inside the content frame. It lives in
 * the layout rather than on a page, so a session survives navigation.
 */
export default function TerminalManager() {
  const sessions = useTerminalStore((s) => s.sessions);
  const minimizeAll = useTerminalStore((s) => s.minimizeAll);
  const reconnectTarget = useTerminalStore((s) => s.reconnectTarget);
  const tenantId = useAuthStore((s) => s.tenant) ?? "";
  const { namespaces } = useNamespaces();
  const workspace = useWorkspaceTabs();
  const enteringRef = useRef<string | null>(null);

  const [connectTarget, setConnectTarget] = useState<{
    uid: string;
    name: string;
    sshid: string;
  } | null>(null);

  useEffect(() => {
    if (!reconnectTarget) return;
    const home = reconnectTarget.tenant;
    if (home && home !== tenantId) {
      if (enteringRef.current === home) return;
      const homeNamespace = namespaces.find((ns) => ns.tenant_id === home);
      if (!homeNamespace) {
        useTerminalStore.getState().clearReconnect();
        return;
      }
      enteringRef.current = home;
      void workspace
        .openNamespace(home, homeNamespace.name)
        .then((entered) => {
          enteringRef.current = null;
          if (!entered) useTerminalStore.getState().clearReconnect();
        });
      return;
    }
    useTerminalStore.getState().clearReconnect();
    const nsName = namespaces.find((ns) => ns.tenant_id === tenantId)?.name;
    const sshid = nsName
      ? buildSshid(nsName, reconnectTarget.deviceName)
      : reconnectTarget.deviceUid;
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setConnectTarget({
      uid: reconnectTarget.deviceUid,
      name: reconnectTarget.deviceName,
      sshid,
    });
  }, [reconnectTarget, tenantId, namespaces, workspace]);

  const location = useLocation();
  const prevPathRef = useRef(location.pathname);
  useEffect(() => {
    if (location.pathname !== prevPathRef.current) {
      prevPathRef.current = location.pathname;
      if (!useTerminalStore.getState().dockPendingRestore()) minimizeAll();
    }
  }, [location.pathname, minimizeAll]);

  return (
    <>
      {connectTarget && (
        <ConnectDrawer
          open
          onClose={() => setConnectTarget(null)}
          deviceUid={connectTarget.uid}
          deviceName={connectTarget.name}
          sshid={connectTarget.sshid}
        />
      )}

      {sessions.map((s) => {
        const isVisible = s.state !== "minimized";

        return (
          <div
            key={s.id}
            className={cn(
              "absolute inset-0 z-terminal flex flex-col bg-background",
              "transition-[opacity,transform] duration-200 ease-out",
              isVisible
                ? "opacity-100 translate-y-0"
                : "opacity-0 translate-y-3 pointer-events-none",
            )}
          >
            <TerminalInstance session={s} visible={isVisible} />
          </div>
        );
      })}

      <RecordingSnackbar />
    </>
  );
}
