import { useCallback, useState } from "react";
import {
  useTerminalStore,
  type TerminalSession,
} from "../../stores/terminalStore";
import { useTerminalThemeStore } from "../../stores/terminalThemeStore";
import NamespaceSelector from "./NamespaceSelector";

import UserMenu from "./UserMenu";
import { TerminalInfo, TerminalActions } from "../terminal/TerminalControls";

export default function AppBar() {
  const activeSession = useTerminalStore((s) =>
    s.sessions.find((s) => s.state !== "minimized"),
  );
  const themeBg = useTerminalThemeStore((s) => s.theme.colors.background);

  const [displayed, setDisplayed] = useState<TerminalSession | null>(
    activeSession ?? null,
  );
  const [phase, setPhase] = useState<"idle" | "fading-out" | "swapped">("idle");
  const [pending, setPending] = useState<TerminalSession | null>(null);
  const [trackedId, setTrackedId] = useState<string | undefined>(
    activeSession?.id,
  );

  const visible = phase === "idle";
  const currentId = activeSession?.id;

  // Detect changes during render (React-blessed setState-during-render pattern)
  if (currentId !== trackedId) {
    setTrackedId(currentId);

    if (phase === "fading-out" || phase === "swapped") {
      // Already animating — stash the latest value
      setPending(activeSession ?? null);
    } else if (!!activeSession !== !!displayed) {
      // Mode changed (terminal ↔ namespace) — start crossfade
      setPending(activeSession ?? null);
      setPhase("fading-out");
    } else if (activeSession) {
      // Same mode, different session — instant swap
      setDisplayed(activeSession);
    }
  }

  const handleTransitionEnd = useCallback(() => {
    if (phase !== "fading-out") return;

    // Fade-out finished — swap content, then fade in after one paint frame
    setDisplayed(pending);
    setPending(null);
    setPhase("swapped");
    requestAnimationFrame(() => {
      requestAnimationFrame(() => setPhase("idle"));
    });
  }, [phase, pending]);

  return (
    <header
      className={`relative z-10 h-14 border-b px-5 flex items-center justify-between shrink-0 transition-colors duration-300 ${
        displayed ? "border-transparent" : "bg-surface border-border"
      }`}
      style={displayed ? { backgroundColor: themeBg } : undefined}
    >
      {/* Left: crossfade with vertical slide */}
      <div
        onTransitionEnd={handleTransitionEnd}
        className={`min-w-0 transition-all duration-150 ease-out ${visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-2"}`}
      >
        {displayed ? (
          <TerminalInfo session={displayed} />
        ) : (
          <NamespaceSelector />
        )}
      </div>

      <div className="flex items-center gap-1">
        {/* Right: terminal actions fade + slide */}
        <div
          className={`flex items-center transition-all duration-150 ease-out ${
            displayed && visible
              ? "opacity-100 translate-y-0"
              : "opacity-0 translate-y-2 pointer-events-none"
          }`}
        >
          {displayed && <TerminalActions session={displayed} />}
        </div>

        <UserMenu />
      </div>
    </header>
  );
}
