import { useEffect, useRef, useState } from "react";
import { useLocation } from "react-router-dom";
import { useTerminalStore } from "../../stores/terminalStore";
import { useNamespacesStore } from "../../stores/namespacesStore";
import ConnectDrawer from "../ConnectDrawer";
import TerminalInstance from "./TerminalInstance";
import TerminalTaskbar from "./TerminalTaskbar";

export default function TerminalManager() {
  const sessions = useTerminalStore((s) => s.sessions);
  const minimizeAll = useTerminalStore((s) => s.minimizeAll);
  const reconnectTarget = useTerminalStore((s) => s.reconnectTarget);
  const currentNamespace = useNamespacesStore((s) => s.currentNamespace);

  const [connectTarget, setConnectTarget] = useState<{
    uid: string;
    name: string;
    sshid: string;
  } | null>(null);

  // Open ConnectDrawer when a reconnect is requested (works from any page)
  useEffect(() => {
    if (!reconnectTarget) return;
    useTerminalStore.getState().clearReconnect();
    const nsName = currentNamespace?.name;
    const sshid = nsName
      ? `${nsName}.${reconnectTarget.deviceName}@${nsName}`
      : reconnectTarget.deviceUid;
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setConnectTarget({
      uid: reconnectTarget.deviceUid,
      name: reconnectTarget.deviceName,
      sshid,
    });
  }, [reconnectTarget, currentNamespace]);

  // Auto-minimize terminal when navigating to another page
  const location = useLocation();
  const prevPathRef = useRef(location.pathname);
  useEffect(() => {
    if (location.pathname !== prevPathRef.current) {
      prevPathRef.current = location.pathname;
      minimizeAll();
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
        const isFullscreen = s.state === "fullscreen";

        return (
          <div
            key={s.id}
            className={[
              "fixed top-14 bottom-0 right-0 z-50 flex flex-col bg-background",
              "transition-[opacity,transform,left] duration-200 ease-out",
              isFullscreen ? "left-0" : "left-[220px]",
              isVisible
                ? "opacity-100 translate-y-0"
                : "opacity-0 translate-y-3 pointer-events-none",
            ].join(" ")}
          >
            <TerminalInstance session={s} visible={isVisible} />
          </div>
        );
      })}

      <TerminalTaskbar />
    </>
  );
}
