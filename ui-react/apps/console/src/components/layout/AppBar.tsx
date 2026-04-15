import { useCallback, useState } from "react";
import {
  useTerminalStore,
  type TerminalSession,
} from "@/stores/terminalStore";
import { useTerminalThemeStore } from "@/stores/terminalThemeStore";
import { Bars3Icon } from "@heroicons/react/24/outline";
import NamespaceSelector from "./NamespaceSelector";

import UserMenu from "./UserMenu";
import InvitationsMenu from "./InvitationsMenu";
import { TerminalInfo, TerminalActions } from "../terminal/TerminalControls";

interface AppBarProps {
  onMenuToggle?: () => void;
}

export default function AppBar({ onMenuToggle }: AppBarProps) {
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

  // Safety net: if fade-out gets stuck (e.g. transitionend never fires
  // because the page re-mounts during navigation), skip straight to idle.
  if (phase === "fading-out" && !activeSession && !pending) {
    setDisplayed(null);
    setPhase("idle");
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
      className={`relative z-50 h-14 border-b px-3 sm:px-5 flex items-center justify-between shrink-0 transition-colors duration-300 ${
        displayed ? "border-transparent" : "bg-surface border-border"
      }`}
      style={displayed ? { backgroundColor: themeBg } : undefined}
    >
      {/* Left: menu toggle + crossfade with vertical slide */}
      <div className="flex items-center gap-1 min-w-0">
        {onMenuToggle && (
          <button
            type="button"
            onClick={onMenuToggle}
            className="lg:hidden p-2 -ml-1 rounded-md text-text-muted hover:text-text-primary hover:bg-hover-subtle transition-colors"
            aria-label="Open navigation menu"
          >
            <Bars3Icon className="w-5 h-5" />
          </button>
        )}
        <div
          onTransitionEnd={handleTransitionEnd}
          className={`min-w-0 transition-all duration-150 ease-out ${visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-2"}`}
        >
          {displayed
            ? (
              <TerminalInfo session={displayed} />
            )
            : (
              <NamespaceSelector />
            )}
        </div>
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

        <InvitationsMenu />
        <UserMenu />
      </div>
    </header>
  );
}
