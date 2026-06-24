import { useCallback, useEffect, useReducer, useRef } from "react";
import { useTerminalStore, type TerminalSession } from "@/stores/terminalStore";
import { useTerminalThemeStore } from "@/stores/terminalThemeStore";
import { Bars3Icon } from "@heroicons/react/24/outline";
import { IconButton } from "@shellhub/design-system/primitives";
import NamespaceSelector from "./NamespaceSelector";

import UserMenu from "./UserMenu";
import InvitationsMenu from "./InvitationsMenu";
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

const activeSessionOf = (
  sessions: TerminalSession[],
): TerminalSession | null =>
  sessions.find((s) => s.state !== "minimized") ?? null;

// Pure reducer for the left-content crossfade state machine — safe under
// StrictMode double-invocation.
function crossfadeReducer(
  state: CrossfadeState,
  action: CrossfadeAction,
): CrossfadeState {
  switch (action.type) {
    case "active-changed": {
      const next = action.session;
      if (state.phase !== "idle") {
        // Safety net: active session vanished mid fade-out — settle straight to idle.
        if (state.phase === "fading-out" && !next) {
          return { displayed: null, phase: "idle", pending: null };
        }
        // Already animating — stash the latest value to apply when the fade ends.
        return { ...state, pending: next };
      }
      // Mode change (terminal <-> namespace) — start the crossfade.
      if (!!next !== !!state.displayed) {
        return { ...state, pending: next, phase: "fading-out" };
      }
      // Same mode, different session — instant swap (no flash).
      if (next) {
        return { ...state, displayed: next };
      }
      return state;
    }
    // Fade-out finished — swap in the pending content, then fade back in (see handler).
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

  // The crossfade is driven by store changes — an external event fired OUTSIDE
  // React's render/effect cycle — not detected during render. Dispatching from the
  // subscription callback (never during render, never synchronously in an effect
  // body) keeps both react-hooks/set-state-in-render and set-state-in-effect happy
  // while removing all render-phase state updates — the fix for the audit's
  // concurrent-rendering concern (team#137). dispatch is stable so the effect runs
  // once ([] deps); the listener fires on every store change but early-returns
  // unless the active-session id actually changed.
  useEffect(() => {
    const unsubscribe = useTerminalStore.subscribe((state) => {
      const next = activeSessionOf(state.sessions);
      if (next?.id === prevIdRef.current) return;
      prevIdRef.current = next?.id;
      dispatch({ type: "active-changed", session: next });
    });
    return unsubscribe;
  }, []);

  // Fade-out transition ended — commit the swap, then fade back in over two paint frames.
  const handleTransitionEnd = useCallback(() => {
    if (phase !== "fading-out") return;
    dispatch({ type: "fade-out-done" });
    requestAnimationFrame(() => {
      requestAnimationFrame(() => dispatch({ type: "settle-idle" }));
    });
  }, [phase]);

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
          className={`min-w-0 transition-all duration-150 ease-out ${visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-2"}`}
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
          className={`flex items-center transition-all duration-150 ease-out ${
            displayed && visible
              ? "opacity-100 translate-y-0"
              : "opacity-0 translate-y-2 pointer-events-none"
          }`}
        >
          {displayed && <TerminalActions session={displayed} />}
        </div>

        <InvitationsMenu />
        <SupportButton />
        <UserMenu />
      </div>
    </header>
  );
}
