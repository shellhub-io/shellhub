import type { CSSProperties, ReactNode } from "react";
import { PauseIcon } from "@heroicons/react/24/solid";
import { cn } from "@shellhub/design-system/cn";
import {
  ansiColor,
  useTerminalThemeStore,
  type AnsiColorName,
  type TerminalTheme,
} from "@/stores/terminalThemeStore";
import TerminalSettingsBody from "@/components/terminal/TerminalSettingsBody";
import {
  useSessionPlayerStore,
  type PlayerControls,
} from "@/stores/sessionPlayerStore";
import ScrollFades from "@/components/common/ScrollFades";
import { PlayerBarShell, PlayerTime } from "@/components/sessions/PlayerBar";
import { useScrollEdges } from "@/hooks/useScrollEdges";
import { useIsDesktop } from "@/hooks/useIsDesktop";
import { useDockable } from "@/hooks/useDockable";
import { useStuck } from "@/hooks/useStuck";

function Output({ theme }: { theme: TerminalTheme }) {
  const colored = (name: AnsiColorName, text: ReactNode) => (
    <span style={{ color: ansiColor(theme.colors, name) }}>{text}</span>
  );
  const prompt = (
    <>
      {colored("green", "root@nuc-7")}:{colored("blue", "~")}${" "}
    </>
  );

  return (
    <>
      {prompt}ls -la{"\n"}
      total 36{"\n"}
      drwxr-xr-x 5 root root 4096 Sep 26 10:14 {colored("blue", ".")}
      {"\n"}
      -rw-r--r-- 1 root root 220 Sep 26 09:02 .bashrc{"\n"}
      drwxr-xr-x 2 root root 4096 Sep 26 10:11 {colored("blue", "bin")}
      {"\n"}
      -rwxr-xr-x 1 root root 1843 Sep 26 10:14 {colored("green", "deploy.sh")}
      {"\n"}
      -rw-r--r-- 1 root root 512 Sep 26 09:40 notes.txt{"\n"}
      {prompt}systemctl status shellhub-agent{"\n"}
      {colored("green", "●")} shellhub-agent.service - ShellHub Agent{"\n"}
      {"     "}Active: {colored("green", "active (running)")} since Fri 10:02
      {"\n"}
      {"   "}Warnings: {colored("yellow", "1")} Errors: {colored("red", "0")}
      {"\n"}
      {prompt}
      <span
        aria-hidden="true"
        className="inline-block w-[0.6em] animate-pulse"
        style={{ background: theme.colors.cursor ?? theme.colors.foreground }}
      >
        {" "}
      </span>
    </>
  );
}

function PreviewPlayerBar({ mode }: { mode: PlayerControls }) {
  if (mode === "hidden") return null;
  return (
    <PlayerBarShell
      aria-hidden="true"
      className={cn(
        "shadow-none",
        mode === "auto" && "opacity-0 group-hover:opacity-100",
      )}
    >
      <span className="shrink-0 w-7 h-7 flex items-center justify-center rounded-full bg-primary text-white">
        <PauseIcon className="w-3 h-3" />
      </span>
      <PlayerTime>01:24 / 04:10</PlayerTime>
      <span className="relative flex-1 h-full">
        <span className="absolute inset-x-0 top-1/2 -translate-y-1/2 h-[3px] rounded-full bg-border">
          <span className="block h-full w-[34%] rounded-full bg-primary" />
        </span>
        <span className="absolute top-1/2 left-[34%] w-[11px] h-[11px] -translate-x-1/2 -translate-y-1/2 rounded-full bg-primary border-2 border-surface" />
      </span>
      <span className="shrink-0 w-9 text-center text-xs font-mono tabular-nums text-text-secondary">
        1×
      </span>
    </PlayerBarShell>
  );
}

function Preview({
  theme,
  fontFamily,
  fontSize,
  playerControls,
}: {
  theme: TerminalTheme;
  fontFamily: string;
  fontSize: number;
  playerControls: PlayerControls;
}) {
  return (
    <div
      style={{ background: theme.colors.background }}
      className="group relative h-full min-w-0 px-3 py-2 transition-colors duration-200"
    >
      <div className="h-full overflow-hidden">
        <pre
          aria-label="Terminal preview"
          style={{
            color: theme.colors.foreground,
            fontFamily,
            fontSize,
            lineHeight: "normal",
          }}
          className="m-0 whitespace-pre"
        >
          <Output theme={theme} />
        </pre>
      </div>
      <PreviewPlayerBar mode={playerControls} />
    </div>
  );
}

const PREVIEW_HEIGHT = "h-40";

/**
 * The terminal settings beside a preview of the terminal they change, drawn as the console draws a
 * terminal's contents. On a wide window the two fill the page down to its end and the settings
 * scroll inside theirs, so the page itself never scrolls and the preview never leaves view. On a
 * narrow one the preview sits above the settings and stays pinned as they scroll; dragged out of
 * its place it floats where it is let go and the place closes behind it, and dragged back over
 * the place it opens again for the preview to settle in. The preview is static output, not a
 * session.
 */
export default function TerminalSettingsPreview() {
  const {
    theme: shown,
    fontFamilyWithFallback,
    fontSize,
  } = useTerminalThemeStore();
  const playerControls = useSessionPlayerStore((s) => s.controls);
  const {
    ref: scrollRef,
    moreAbove,
    moreBelow,
  } = useScrollEdges<HTMLDivElement>();
  const narrow = !useIsDesktop();
  const {
    slot,
    placement,
    dragging,
    overDock,
    slotOpen,
    frame,
    onDocked,
    onDockInterrupted,
    handlers,
  } = useDockable(narrow);
  const { sentinel, stuck } = useStuck();
  const detached = placement !== "docked";
  const style = {
    "--float-top": `${frame.top}px`,
    "--float-left": `${frame.left}px`,
    "--float-width": `${frame.width}px`,
  } as CSSProperties;

  const previewBox = (
    <div
      {...handlers}
      onTransitionEnd={onDocked}
      onTransitionCancel={onDockInterrupted}
      style={style}
      data-floating={narrow && (detached || stuck)}
      data-dragging={dragging}
      data-dark={shown.dark}
      className={cn(
        PREVIEW_HEIGHT,
        "z-raised touch-none cursor-grab active:cursor-grabbing select-none rounded-xl border border-black/15 data-[dark=true]:border-white/10 overflow-clip",
        "transition-[box-shadow,transform] duration-200",
        "data-[dragging=true]:scale-[1.03]",
        "data-[floating=true]:shadow-float",
        detached
          ? "fixed top-[var(--float-top)] left-[var(--float-left)] w-[var(--float-width)]"
          : "sticky top-2 mb-4 lg:mb-0 lg:static lg:h-auto lg:min-h-0 lg:touch-auto lg:cursor-auto lg:border-border",
        placement === "docking" &&
          "transition-[top,box-shadow,transform] duration-300 ease-out",
      )}
    >
      <Preview
        theme={shown}
        fontFamily={fontFamilyWithFallback}
        fontSize={fontSize}
        playerControls={playerControls}
      />
      <span
        aria-hidden="true"
        style={{ background: shown.colors.foreground }}
        className="lg:hidden absolute bottom-1.5 left-1/2 -translate-x-1/2 w-10 h-1 rounded-full opacity-50"
      />
    </div>
  );

  return (
    <div className="relative grid grid-cols-[minmax(0,1fr)] items-start lg:gap-4 lg:flex-1 lg:min-h-0 lg:items-stretch lg:grid-cols-[minmax(0,1fr)_18rem] lg:grid-rows-[minmax(0,1fr)]">
      <div ref={sentinel} aria-hidden="true" className="absolute top-0 h-px" />
      {detached && (
        <div
          ref={slot}
          aria-hidden="true"
          data-open={overDock}
          className={cn(
            "relative rounded-xl overflow-clip opacity-0 transition-all duration-300 ease-out",
            slotOpen ? [PREVIEW_HEIGHT, "mb-4"] : "h-0 mb-0",
            "data-[open=true]:opacity-100",
          )}
        >
          <span
            style={{ background: shown.colors.background }}
            className="absolute inset-0 opacity-40"
          />
        </div>
      )}
      {previewBox}
      <div className="relative lg:min-h-0 flex flex-col rounded-xl border border-border bg-card overflow-clip">
        <div
          ref={scrollRef}
          className="relative lg:absolute lg:inset-0 lg:overflow-y-auto overscroll-contain"
        >
          <TerminalSettingsBody />
        </div>
        <ScrollFades moreAbove={moreAbove} moreBelow={moreBelow} tone="card" />
      </div>
    </div>
  );
}
