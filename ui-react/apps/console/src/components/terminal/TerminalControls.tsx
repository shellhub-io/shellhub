import { useState } from "react";
import { createPortal } from "react-dom";
import {
  XMarkIcon,
  Cog6ToothIcon,
  MinusIcon,
} from "@heroicons/react/24/outline";
import { useTerminalStore } from "@/stores/terminalStore";
import type { TerminalSession } from "@/stores/terminalStore";
import TerminalSettingsDrawer from "./TerminalSettingsDrawer";

/** Terminal info shown on the left side of the AppBar */
export function TerminalInfo({ session }: { session: TerminalSession }) {
  const status = useTerminalStore(
    (s) =>
      s.sessions.find((ss) => ss.id === session.id)?.connectionStatus
      ?? "disconnected",
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
            onClick={() => toggleFullscreen(session.id)}
            className="w-3.5 h-3.5 rounded-full bg-[#28c840] border border-[#1aab29] flex items-center justify-center transition-all hover:brightness-110 active:brightness-90"
            title={isFullscreen ? "Exit fullscreen" : "Fullscreen"}
          >
            <svg
              className="w-2 h-2 text-[#006500] opacity-0 group-hover/lights:opacity-100 transition-opacity"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth={3}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d={isFullscreen
                  ? "M8 4v4H4M20 8h-4V4M16 20v-4h4M4 16h4v4"
                  : "M4 8V4h4M16 4h4v4M20 16v4h-4M8 20H4v-4"}
              />
            </svg>
          </button>
        </div>

        {/* Settings */}
        <button
          onClick={() => setSettingsOpen(true)}
          className="p-1.5 rounded-md text-white/30 hover:text-white/60 transition-colors"
          title="Terminal Settings"
        >
          <Cog6ToothIcon className="w-4 h-4" />
        </button>
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
