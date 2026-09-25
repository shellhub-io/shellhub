import {
  useEffect,
  useEffectEvent,
  useRef,
  useState,
  type CSSProperties,
  type PointerEvent,
} from "react";
import { create, type Player } from "asciinema-player";
import "asciinema-player/dist/bundle/asciinema-player.css";
import { PlayIcon, PauseIcon } from "@heroicons/react/24/solid";
import { cn } from "@shellhub/design-system/cn";
import { Dropdown, IconButton } from "@shellhub/design-system/primitives";
import {
  ansiPalette,
  useTerminalThemeStore,
  type TerminalThemeColors,
} from "@/stores/terminalThemeStore";
import { useSettledValue } from "@/hooks/useSettledValue";
import {
  nextPlayerControls,
  useSessionPlayerStore,
} from "@/stores/sessionPlayerStore";
import { useIdleControls } from "./useIdleControls";

const SPEEDS = [0.5, 1, 1.5, 2, 4] as const;
type Speed = (typeof SPEEDS)[number];

function castPalette(colors: TerminalThemeColors): CSSProperties {
  return Object.fromEntries([
    ["--cast-foreground", colors.foreground],
    ["--cast-background", colors.background],
    ...ansiPalette(colors).map((color, i) => [`--cast-${i}`, color]),
  ]) as CSSProperties;
}

const NOTICE_MS = 1200;

const LOOK_SETTLE_MS = 250;

function padZero(num: number): string {
  return num.toString().padStart(2, "0");
}

function formatTime(secs: number, showHours = false): string {
  const h = padZero(Math.floor(secs / 3600));
  const m = padZero(Math.floor(secs / 60) % 60);
  const s = padZero(Math.floor(secs % 60));

  if (showHours || h !== "00") return `${h}:${m}:${s}`;
  return `${m}:${s}`;
}

function KeyboardIcon({ className }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={1.5}
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
    >
      <rect x="2" y="7" width="20" height="11" rx="2" />
      <path d="M6 11h.01M10 11h.01M14 11h.01M18 11h.01M8 15h8" />
    </svg>
  );
}

const SHORTCUTS = [
  { keys: ["Space"], description: "Pause / resume" },
  { keys: ["←", "→"], description: "Rewind / fast-forward 5s" },
  { keys: ["Shift+←", "Shift+→"], description: "Rewind / fast-forward 10%" },
  { keys: ["0–9"], description: "Jump to 0%, 10%, … 90%" },
  { keys: [",", "."], description: "Step back / forward one frame (paused)" },
  { keys: ["F"], description: "Toggle fullscreen" },
  { keys: ["H"], description: "Controls: auto-hide, always, hidden" },
  { keys: ["Esc"], description: "Exit fullscreen / hide controls" },
];

interface SessionPlayerProps {
  logs: string;
  visible?: boolean;
}

function Timeline({
  currentTime,
  duration,
  onSeek,
}: {
  currentTime: number;
  duration: number;
  onSeek: (t: number) => void;
}) {
  const [hover, setHover] = useState<number | null>(null);
  const progress = duration > 0 ? Math.min(1, currentTime / duration) : 0;

  const fractionAt = (e: PointerEvent<HTMLDivElement>) => {
    const box = e.currentTarget.getBoundingClientRect();
    return Math.max(0, Math.min(1, (e.clientX - box.left) / box.width));
  };

  return (
    <div
      role="slider"
      tabIndex={0}
      aria-label="Seek"
      aria-valuemin={0}
      aria-valuemax={Math.round(duration)}
      aria-valuenow={Math.round(currentTime)}
      aria-valuetext={formatTime(currentTime, duration >= 3600)}
      aria-disabled={duration <= 0 || undefined}
      className="relative flex-1 h-full cursor-pointer touch-none outline-none focus-visible:ring-1 focus-visible:ring-primary/50 rounded"
      onPointerDown={(e) => {
        if (duration <= 0) return;
        e.currentTarget.setPointerCapture(e.pointerId);
        onSeek(duration * fractionAt(e));
      }}
      onPointerMove={(e) => {
        setHover(fractionAt(e));
        if (e.buttons & 1 && duration > 0) onSeek(duration * fractionAt(e));
      }}
      onPointerLeave={() => setHover(null)}
    >
      <div className="absolute inset-x-0 top-1/2 -translate-y-1/2 h-[3px] rounded-full bg-border">
        <div
          className="h-full rounded-full bg-primary"
          style={{ width: `${progress * 100}%` }}
        />
      </div>
      {duration > 0 && (
        <div
          className="absolute top-1/2 w-[11px] h-[11px] -translate-x-1/2 -translate-y-1/2 rounded-full bg-primary border-2 border-surface"
          style={{ left: `${progress * 100}%` }}
        />
      )}
      {hover !== null && duration > 0 && (
        <div
          className="pointer-events-none absolute bottom-full mb-2 -translate-x-1/2 px-2 py-1 rounded-md border border-border bg-surface text-2xs font-mono tabular-nums text-text-secondary whitespace-nowrap"
          style={{ left: `${hover * 100}%` }}
        >
          {formatTime(duration * hover, duration >= 3600)}
        </div>
      )}
    </div>
  );
}

/**
 * Replays a recorded session with asciinema, in the terminal's own colours, font and size, so a
 * recording looks like the terminal it came from. The recording is passed in already fetched,
 * since it is large and only wanted once the player is opened. While visible is false, as for a
 * recording whose tab is in the background, it pauses and ignores the keyboard. Its shortcuts
 * answer only when focus is in the player or on nothing, never with Ctrl, Alt or Meta held.
 *
 * The controls float over the bottom of the screen. By default they fade after two seconds
 * without pointer or keyboard activity while the recording plays, and stay while it is paused,
 * while the pointer is on them, while they hold keyboard focus or while the shortcuts are open.
 * Esc puts them away until the next activity, and H cycles the session player store's
 * preference: auto-hide, always shown, hidden. The shortcuts work whatever the controls show.
 */
export default function SessionPlayer({
  logs,
  visible = true,
}: SessionPlayerProps) {
  const rootRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const playerRef = useRef<Player | null>(null);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const endedRef = useRef(false);
  const currentTimeRef = useRef(0);
  const durationRef = useRef(0);
  const isPlayingRef = useRef(false);
  const wantsPlayRef = useRef(true);
  const resumeRef = useRef({ at: 0, playing: true });

  const [isPlaying, setIsPlaying] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);
  const [speed, setSpeed] = useState<Speed>(1);
  const [showShortcuts, setShowShortcuts] = useState(false);
  const [notice, setNotice] = useState<string | null>(null);
  const noticeRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const colors = useSettledValue(
    useTerminalThemeStore((s) => s.theme.colors),
    LOOK_SETTLE_MS,
  );
  const controls = useSessionPlayerStore((s) => s.controls);
  const setControls = useSessionPlayerStore((s) => s.setControls);
  const fontFamily = useSettledValue(
    useTerminalThemeStore((s) => s.fontFamilyWithFallback),
    LOOK_SETTLE_MS,
  );
  const fontSize = useSettledValue(
    useTerminalThemeStore((s) => s.fontSize),
    LOOK_SETTLE_MS,
  );
  const [keyboardInBar, setKeyboardInBar] = useState(false);
  const idle = useIdleControls();
  const onPlaybackStart = useEffectEvent(() => {
    if (visible) idle.wake();
    return visible;
  });

  const stopClock = () => {
    if (timerRef.current !== null) clearInterval(timerRef.current);
    timerRef.current = null;
  };

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;
    const resume = resumeRef.current;
    const player = create({ data: logs }, container, {
      fit: false,
      controls: false,
      speed,
      startAt: resume.at,
      terminalFontFamily: fontFamily,
      terminalFontSize: `${fontSize}px`,
    });
    playerRef.current = player;

    player.addEventListener("playing", () => {
      if (!onPlaybackStart()) {
        wantsPlayRef.current = false;
        void player.pause();
        return;
      }
      endedRef.current = false;
      isPlayingRef.current = true;
      setIsPlaying(true);
      stopClock();
      timerRef.current = setInterval(() => {
        const t = player.getCurrentTime();
        if (t != null) {
          currentTimeRef.current = t;
          setCurrentTime(t);
        }
      }, 100);
      const d = player.getDuration();
      if (d != null) {
        durationRef.current = d;
        setDuration(d);
      }
    });

    player.addEventListener("ended", () => {
      endedRef.current = true;
      isPlayingRef.current = false;
      wantsPlayRef.current = false;
      setIsPlaying(false);
      stopClock();
    });

    if (resume.playing) void player.play();
    else void player.seek(resume.at);

    return () => {
      resumeRef.current = {
        at: endedRef.current ? 0 : currentTimeRef.current,
        playing: wantsPlayRef.current,
      };
      stopClock();
      player.dispose();
      playerRef.current = null;
    };
  }, [logs, speed, fontFamily, fontSize, colors]);

  const seekTo = (t: number) => {
    const wasPlaying = isPlayingRef.current;
    const clamped = Math.max(0, Math.min(durationRef.current, t));
    if (clamped < durationRef.current) endedRef.current = false;
    void playerRef.current?.pause();
    void playerRef.current?.seek(clamped).then(() => {
      if (wasPlaying) void playerRef.current?.play();
    });
    currentTimeRef.current = clamped;
    setCurrentTime(clamped);
  };

  const handlePlayPause = () => {
    if (isPlayingRef.current) {
      void playerRef.current?.pause();
      isPlayingRef.current = false;
      wantsPlayRef.current = false;
      setIsPlaying(false);
      stopClock();
    } else {
      void playerRef.current?.play();
      isPlayingRef.current = true;
      wantsPlayRef.current = true;
      setIsPlaying(true);
    }
  };

  const announce = (message: string) => {
    setNotice(message);
    if (noticeRef.current !== null) clearTimeout(noticeRef.current);
    noticeRef.current = setTimeout(() => setNotice(null), NOTICE_MS);
  };

  useEffect(
    () => () => {
      if (noticeRef.current !== null) clearTimeout(noticeRef.current);
    },
    [],
  );

  const handleKeyDown = useEffectEvent((e: KeyboardEvent) => {
    if (!visible || e.defaultPrevented) return;
    if (e.ctrlKey || e.metaKey || e.altKey) return;
    const target = e.target instanceof Element ? e.target : null;
    const onPage = target === null || target === document.body;
    if (!onPage && !rootRef.current?.contains(target)) return;
    if (
      (e.key === " " || e.key === "Enter") &&
      target?.closest("button, [role=button]")
    )
      return;

    idle.wake();
    switch (e.key) {
      case " ":
        e.preventDefault();
        handlePlayPause();
        break;
      case "ArrowLeft":
        e.preventDefault();
        seekTo(
          currentTimeRef.current - (e.shiftKey ? durationRef.current * 0.1 : 5),
        );
        break;
      case "ArrowRight":
        e.preventDefault();
        seekTo(
          currentTimeRef.current + (e.shiftKey ? durationRef.current * 0.1 : 5),
        );
        break;
      case ",":
        if (!isPlayingRef.current) seekTo(currentTimeRef.current - 0.1);
        break;
      case ".":
        if (!isPlayingRef.current) seekTo(currentTimeRef.current + 0.1);
        break;
      case "f":
      case "F":
        if (document.fullscreenElement) {
          void document.exitFullscreen();
        } else {
          void rootRef.current?.requestFullscreen();
        }
        break;
      case "Escape":
        if (document.fullscreenElement) void document.exitFullscreen();
        else {
          setShowShortcuts(false);
          idle.stow();
        }
        return;
      case "h":
      case "H": {
        const next = nextPlayerControls(controls);
        setControls(next.value);
        announce(next.notice);
        break;
      }
      default:
        if (e.key >= "0" && e.key <= "9" && !e.shiftKey) {
          seekTo((durationRef.current * parseInt(e.key)) / 10);
        }
    }
  });

  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => handleKeyDown(e);
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, []);

  const pauseWhenHidden = useEffectEvent(() => {
    if (isPlayingRef.current) handlePlayPause();
    wantsPlayRef.current = false;
  });
  useEffect(() => {
    if (!visible) pauseWhenHidden();
  }, [visible]);

  const nextSpeed = SPEEDS[(SPEEDS.indexOf(speed) + 1) % SPEEDS.length];

  const barShown = idle.shown(
    controls === "always" || !isPlaying || showShortcuts || keyboardInBar,
  );

  return (
    <div
      ref={rootRef}
      className="cast-palette relative h-full"
      style={{ ...castPalette(colors), backgroundColor: colors.background }}
      onPointerMove={idle.wake}
      onPointerDown={idle.wake}
    >
      <div
        className={cn(
          "absolute inset-0 overflow-auto p-4 [scrollbar-gutter:stable]",
          controls !== "hidden" && "pb-24",
        )}
      >
        <div ref={containerRef} />
      </div>

      {controls !== "hidden" && (
        <div
          className={cn(
            "absolute bottom-[18px] left-1/2 -translate-x-1/2 w-[min(calc(100%-32px),660px)] h-[54px] flex items-center gap-2.5 pl-2.5 pr-2 rounded-xl border border-border bg-surface/90 backdrop-blur-sm shadow-lg transition-opacity duration-150",
            barShown ? "opacity-100" : "opacity-0",
          )}
          onFocus={(e) => setKeyboardInBar(e.target.matches(":focus-visible"))}
          onBlur={(e) => {
            if (!e.currentTarget.contains(e.relatedTarget))
              setKeyboardInBar(false);
          }}
          data-testid="player-controls"
          data-state={barShown ? "shown" : "hidden"}
          {...idle.barProps}
        >
          <button
            type="button"
            aria-label={isPlaying ? "Pause" : "Play"}
            onClick={handlePlayPause}
            className="shrink-0 w-7 h-7 flex items-center justify-center rounded-full bg-primary text-white hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/50"
          >
            {isPlaying ? (
              <PauseIcon className="w-3 h-3" />
            ) : (
              <PlayIcon className="w-3 h-3 ml-px" />
            )}
          </button>

          <span className="shrink-0 min-w-[84px] text-center text-xs font-mono tabular-nums text-text-secondary">
            {formatTime(currentTime, duration >= 3600)} / {formatTime(duration)}
          </span>

          <Timeline
            currentTime={currentTime}
            duration={duration}
            onSeek={seekTo}
          />

          <button
            type="button"
            aria-label={`Playback speed ${speed}×, switch to ${nextSpeed}×`}
            title="Playback speed"
            onClick={() => setSpeed(nextSpeed)}
            className="shrink-0 w-9 py-1 rounded text-xs font-mono tabular-nums text-text-secondary hover:text-text-primary hover:bg-hover-subtle focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-primary/50"
          >
            {speed}×
          </button>

          <Dropdown
            mode="content"
            placement="top-end"
            open={showShortcuts}
            onOpenChange={setShowShortcuts}
          >
            <Dropdown.Trigger>
              <IconButton
                variant={showShortcuts ? "primary" : "ghost"}
                className={
                  showShortcuts
                    ? "bg-primary/10 text-primary border border-primary/20"
                    : "border border-transparent"
                }
                aria-label="Keyboard shortcuts"
                title="Keyboard shortcuts"
              >
                <KeyboardIcon className="w-4 h-4" />
              </IconButton>
            </Dropdown.Trigger>

            <Dropdown.Panel
              aria-label="Keyboard shortcuts"
              className="p-3 w-80"
            >
              <p className="text-2xs font-mono font-semibold uppercase tracking-widest text-text-muted/60 mb-2.5">
                Keyboard Shortcuts
              </p>
              <div className="space-y-1.5">
                {SHORTCUTS.map(({ keys, description }) => (
                  <div
                    key={description}
                    className="flex items-center justify-between gap-4"
                  >
                    <span className="text-xs text-text-secondary">
                      {description}
                    </span>
                    <div className="flex items-center gap-1 shrink-0">
                      {keys.map((k) => (
                        <kbd
                          key={k}
                          className="px-1.5 py-0.5 text-2xs font-mono bg-surface border border-border rounded text-text-secondary"
                        >
                          {k}
                        </kbd>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            </Dropdown.Panel>
          </Dropdown>
        </div>
      )}

      {notice && (
        <div
          role="status"
          className="pointer-events-none absolute top-4 left-1/2 -translate-x-1/2 px-3 py-1.5 rounded-lg border border-border bg-surface/90 backdrop-blur-sm text-xs text-text-secondary shadow-lg"
        >
          {notice}
        </div>
      )}
    </div>
  );
}
