import { useState } from "react";
import { createPortal } from "react-dom";
import {
  XMarkIcon,
  Cog6ToothIcon,
  MinusIcon,
  ArrowsPointingOutIcon,
  ArrowsPointingInIcon,
} from "@heroicons/react/24/outline";
import { IconButton } from "@shellhub/design-system/primitives";
import { useTerminalStore } from "@/stores/terminalStore";
import type { TerminalSession } from "@/stores/terminalStore";
import TerminalSettingsDrawer from "./TerminalSettingsDrawer";

const fullscreenButtonIconClassName =
  "w-2 h-2 text-[#006500] opacity-0 group-hover/lights:opacity-100 transition-opacity";

/** Terminal info shown on the left side of the AppBar */
export function TerminalInfo({ session }: { session: TerminalSession }) {
  const status = useTerminalStore(
    (s) =>
      s.sessions.find((ss) => ss.id === session.id)?.connectionStatus ??
      "disconnected",
  );

  return (
    <div className="flex items-center gap-2.5 min-w-0">
      <span
        className={`shrink-0 w-2 h-2 rounded-full transition-colors duration-300 ${
          status === "connected"
            ? "bg-accent-green shadow-[0_0_6px_rgba(130,165,104,0.6)]"
            : status === "connecting"
              ? "bg-accent-yellow animate-pulse"
              : "bg-accent-red"
        }`}
      />
      <span className="text-[13px] font-mono text-text-secondary truncate">
        {status === "connected"
          ? `${session.username}@${session.deviceName}`
          : status === "connecting"
            ? `Connecting to ${session.deviceName}...`
            : `${session.deviceName} — Disconnected`}
      </span>
    </div>
  );
}

/** Terminal action buttons shown on the right side of the AppBar */
export function TerminalActions({ session }: { session: TerminalSession }) {
  const { minimize, toggleFullscreen, close } = useTerminalStore();
  const [settingsOpen, setSettingsOpen] = useState(false);
  const isFullscreen = session.state === "fullscreen";

  return (
    <>
      <div className="flex items-center gap-1">
        {/* macOS-style traffic light controls */}
        <div className="flex items-center gap-2 ml-1.5 group/lights">
          <button
            type="button"
            onClick={() => close(session.id)}
            className="w-3.5 h-3.5 rounded-full bg-[#ff5f57] border border-[#e0443e] flex items-center justify-center transition-all hover:brightness-110 active:brightness-90"
            title="Close"
          >
            <XMarkIcon
              className="w-2 h-2 text-[#4a0002] opacity-0 group-hover/lights:opacity-100 transition-opacity"
              strokeWidth={3}
            />
          </button>
          <button
            type="button"
            onClick={() => minimize(session.id)}
            className="w-3.5 h-3.5 rounded-full bg-[#febc2e] border border-[#dea123] flex items-center justify-center transition-all hover:brightness-110 active:brightness-90"
            title="Minimize"
          >
            <MinusIcon
              className="w-2 h-2 text-[#5a3b00] opacity-0 group-hover/lights:opacity-100 transition-opacity"
              strokeWidth={3}
            />
          </button>
          <button
            type="button"
            onClick={() => toggleFullscreen(session.id)}
            className="w-3.5 h-3.5 rounded-full bg-[#28c840] border border-[#1aab29] flex items-center justify-center transition-all hover:brightness-110 active:brightness-90"
            aria-label={isFullscreen ? "Exit fullscreen" : "Fullscreen"}
            title={isFullscreen ? "Exit fullscreen" : "Fullscreen"}
          >
            {isFullscreen ? (
              <ArrowsPointingInIcon className={fullscreenButtonIconClassName} />
            ) : (
              <ArrowsPointingOutIcon className={fullscreenButtonIconClassName} />
            )}
          </button>
        </div>

        {/* Settings */}
        <IconButton
          title="Terminal settings"
          aria-label="Terminal settings"
          onClick={() => setSettingsOpen(true)}
          className="text-white/30 hover:text-white/60"
        >
          <Cog6ToothIcon className="w-4 h-4" />
        </IconButton>
      </div>

      {createPortal(
        <TerminalSettingsDrawer
          open={settingsOpen}
          onClose={() => setSettingsOpen(false)}
        />,
        document.body,
      )}
    </>
  );
}
