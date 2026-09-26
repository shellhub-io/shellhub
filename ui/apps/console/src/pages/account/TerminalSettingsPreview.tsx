import { useState, type ReactNode } from "react";
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
      className={cn(mode === "auto" && "opacity-0 group-hover:opacity-100")}
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

const SHADOW_ON_DARK = "-18px 0 44px -10px rgba(0, 0, 0, 0.85)";
const SHADOW_ON_LIGHT = "-14px 0 32px -14px rgba(15, 23, 42, 0.35)";

/**
 * The terminal settings beside the terminal they change: a preview drawn as the console draws a
 * terminal's contents, and the Terminal Settings drawer's body docked next to it. Pointing at or
 * focusing a theme shows it in the preview before it is chosen. The preview is static output, not
 * a session.
 */
export default function TerminalSettingsPreview() {
  const { theme, themes, fontFamilyWithFallback, fontSize } =
    useTerminalThemeStore();
  const [preview, setPreview] = useState<string | null>(null);
  const playerControls = useSessionPlayerStore((s) => s.controls);
  const {
    ref: scrollRef,
    moreAbove,
    moreBelow,
  } = useScrollEdges<HTMLDivElement>();
  const shown = themes.find((t) => t.name === preview) ?? theme;

  return (
    <div className="grid grid-cols-[minmax(0,1fr)] lg:grid-cols-[minmax(0,1fr)_20rem] rounded-xl border border-border overflow-clip">
      <div className="h-72 lg:h-[34rem]">
        <Preview
          theme={shown}
          fontFamily={fontFamilyWithFallback}
          fontSize={fontSize}
          playerControls={playerControls}
        />
      </div>
      <div
        style={{ boxShadow: shown.dark ? SHADOW_ON_DARK : SHADOW_ON_LIGHT }}
        className="relative z-raised flex flex-col min-h-0 h-[30rem] lg:h-[34rem] bg-card border-t lg:border-t-0 lg:border-l border-border"
      >
        <div className="relative flex-1 min-h-0 flex flex-col">
          <div
            ref={scrollRef}
            className="relative flex-1 min-h-0 overflow-y-auto overscroll-contain"
          >
            <TerminalSettingsBody onPreviewTheme={setPreview} />
          </div>
          <ScrollFades
            moreAbove={moreAbove}
            moreBelow={moreBelow}
            tone="card"
          />
        </div>
      </div>
    </div>
  );
}
