import { useCallback, useEffect, useRef, useState } from "react";
import { useTerminalStore, type TerminalSession } from "../../stores/terminalStore";
import { useTerminalThemeStore } from "../../stores/terminalThemeStore";
import NamespaceSelector from "./NamespaceSelector";

import UserMenu from "./UserMenu";
import { TerminalInfo, TerminalActions } from "../terminal/TerminalControls";

export default function AppBar() {
  const activeSession = useTerminalStore((s) => s.sessions.find((s) => s.state !== "minimized"));
  const themeBg = useTerminalThemeStore((s) => s.theme.colors.background);

  const [displayed, setDisplayed] = useState<TerminalSession | null>(activeSession ?? null);
  const [visible, setVisible] = useState(true);
  const phaseRef = useRef<"idle" | "fading-out">("idle");
  const pendingRef = useRef<TerminalSession | null | undefined>(undefined);

  useEffect(() => {
    const hasTerminal = !!activeSession;
    const hasDisplayed = !!displayed;

    if (phaseRef.current === "fading-out") {
      // Stash latest value so onTransitionEnd picks it up
      pendingRef.current = activeSession ?? null;
      return;
    }

    if (hasTerminal === hasDisplayed) {
      if (activeSession) setDisplayed(activeSession);
      return;
    }

    // Mode changed: start fade-out, swap happens in onTransitionEnd
    phaseRef.current = "fading-out";
    pendingRef.current = activeSession ?? null;
    setVisible(false);
  }, [activeSession, displayed]);

  const handleTransitionEnd = useCallback(() => {
    if (phaseRef.current !== "fading-out") return;

    // Fade-out finished — swap content and fade in
    phaseRef.current = "idle";
    const next = pendingRef.current;
    pendingRef.current = undefined;
    setDisplayed(next ?? null);
    requestAnimationFrame(() => {
      requestAnimationFrame(() => setVisible(true));
    });
  }, []);

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
        {displayed ? <TerminalInfo session={displayed} /> : <NamespaceSelector />}
      </div>

      <div className="flex items-center gap-1">
        {/* Right: terminal actions fade + slide */}
        <div className={`flex items-center transition-all duration-150 ease-out ${
          displayed && visible
            ? "opacity-100 translate-y-0"
            : "opacity-0 translate-y-2 pointer-events-none"
        }`}>
          {displayed && <TerminalActions session={displayed} />}
        </div>

        <UserMenu />
      </div>
    </header>
  );
}
