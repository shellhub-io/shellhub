import { useCallback, useEffect, useReducer, useRef } from "react";
import { useTerminalStore, type TerminalSession } from "@/stores/terminalStore";
import { useTerminalThemeStore } from "@/stores/terminalThemeStore";
import { Bars3Icon } from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import { IconButton } from "@shellhub/design-system/primitives";
import NamespaceSelector from "./NamespaceSelector";

import UserMenu from "./UserMenu";
import SupportButton from "./SupportButton";
import { TerminalInfo, TerminalActions } from "../terminal/TerminalControls";

interface AppBarProps {
  onMenuToggle?: () => void;
}

type Phase = "idle" | "fading-out" | "swapped";

interface CrossfadeState {
  displayed: TerminalSession | null;
  phase: Phase;
  pending: TerminalSession | null;
}

type CrossfadeAction =
  | { type: "active-changed"; session: TerminalSession | null }
  | { type: "fade-out-done" }
  | { type: "settle-idle" };

const activeSessionOf = (sessions: TerminalSession[]): TerminalSession | null =>
  sessions.find((s) => s.state !== "minimized") ?? null;

function crossfadeReducer(
  state: CrossfadeState,
  action: CrossfadeAction,
): CrossfadeState {
  switch (action.type) {
    case "active-changed": {
      const next = action.session;
      if (state.phase !== "idle") {
        if (state.phase === "fading-out" && !next) {
          return { displayed: null, phase: "idle", pending: null };
        }
        return { ...state, pending: next };
      }
      if (!!next !== !!state.displayed) {
        return { ...state, pending: next, phase: "fading-out" };
      }
      if (next) {
        return { ...state, displayed: next };
      }
      return state;
    }
    case "fade-out-done":
      if (state.phase !== "fading-out") return state;
      return { displayed: state.pending, pending: null, phase: "swapped" };
    case "settle-idle":
      return state.phase === "idle" ? state : { ...state, phase: "idle" };
    default:
      return state;
  }
}

export default function AppBar({ onMenuToggle }: AppBarProps) {
  const themeBg = useTerminalThemeStore((s) => s.theme.colors.background);

  const [{ displayed, phase }, dispatch] = useReducer(
    crossfadeReducer,
    null,
    (): CrossfadeState => ({
      displayed: activeSessionOf(useTerminalStore.getState().sessions),
      phase: "idle",
      pending: null,
    }),
  );

  const prevIdRef = useRef<string | undefined>(
    activeSessionOf(useTerminalStore.getState().sessions)?.id,
  );

  const visible = phase === "idle";

  useEffect(() => {
    const unsubscribe = useTerminalStore.subscribe((state) => {
      const next = activeSessionOf(state.sessions);
      if (next?.id === prevIdRef.current) return;
      prevIdRef.current = next?.id;
      dispatch({ type: "active-changed", session: next });
    });
    return unsubscribe;
  }, []);

  const handleTransitionEnd = useCallback(() => {
    if (phase !== "fading-out") return;
    dispatch({ type: "fade-out-done" });
    requestAnimationFrame(() => {
      requestAnimationFrame(() => dispatch({ type: "settle-idle" }));
    });
  }, [phase]);

  return (
    <header
      className={cn(
        "theme-dark relative z-appbar h-14 border-b px-3 sm:px-5 flex items-center justify-between shrink-0 transition-colors duration-300",
        displayed ? "border-transparent" : "bg-surface border-border",
      )}
      style={displayed ? { backgroundColor: themeBg } : undefined}
    >
      {/* Left: menu toggle + crossfade with vertical slide */}
      <div className="flex items-center gap-1 min-w-0">
        {onMenuToggle && (
          <IconButton
            onClick={onMenuToggle}
            aria-label="Open navigation menu"
            className="lg:hidden -ml-1"
          >
            <Bars3Icon className="w-5 h-5" />
          </IconButton>
        )}
        <div
          onTransitionEnd={handleTransitionEnd}
          className={cn(
            "min-w-0 transition-all duration-150 ease-out",
            visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-2",
          )}
        >
          {displayed ? (
            <TerminalInfo session={displayed} />
          ) : (
            <NamespaceSelector />
          )}
        </div>
      </div>

      <div className="flex items-center gap-1">
        {/* Right: terminal actions fade + slide */}
        <div
          className={cn(
            "flex items-center transition-all duration-150 ease-out",
            displayed && visible
              ? "opacity-100 translate-y-0"
              : "opacity-0 translate-y-2 pointer-events-none",
          )}
        >
          {displayed && <TerminalActions session={displayed} />}
        </div>

        <SupportButton />
        <UserMenu />
      </div>
    </header>
  );
}
