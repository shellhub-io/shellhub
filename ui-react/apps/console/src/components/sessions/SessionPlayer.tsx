import { useEffect, useRef, useState } from "react";
import { create, type AsciinemaPlayer } from "asciinema-player";
import "asciinema-player/dist/bundle/asciinema-player.css";
import { PlayIcon, PauseIcon } from "@heroicons/react/24/solid";

type Speed = 0.5 | 1 | 1.5 | 2;

function formatTime(secs: number): string {
  const m = Math.floor(secs / 60).toString().padStart(2, "0");
  const s = Math.floor(secs % 60).toString().padStart(2, "0");
  return `${m}:${s}`;
}

function KeyboardIcon({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.5} strokeLinecap="round" strokeLinejoin="round" className={className}>
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
  { keys: ["Esc"], description: "Exit fullscreen / close" },
];

interface SessionPlayerProps {
  logs: string;
  onClose?: () => void;
}

export default function SessionPlayer({ logs, onClose }: SessionPlayerProps) {
  const rootRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const playerRef = useRef<AsciinemaPlayer | null>(null);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const speedRef = useRef<Speed>(1);
  const endedRef = useRef(false);
  const currentTimeRef = useRef(0);
  const durationRef = useRef(0);
  const isPlayingRef = useRef(false);

  const [isPlaying, setIsPlaying] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);
  const [speed, setSpeed] = useState<Speed>(1);
  const [showShortcuts, setShowShortcuts] = useState(false);

  const clearTimer = () => {
    if (timerRef.current !== null) {
      clearInterval(timerRef.current);
      timerRef.current = null;
    }
  };

  const startTimer = () => {
    clearTimer();
    timerRef.current = setInterval(() => {
      void playerRef.current?.getCurrentTime().then((t) => {
        currentTimeRef.current = t;
        setCurrentTime(t);
      });
    }, 100);
  };

  const attachListeners = (p: AsciinemaPlayer) => {
    p.addEventListener("playing", () => {
      endedRef.current = false;
      isPlayingRef.current = true;
      setIsPlaying(true);
      startTimer();
      void p.getDuration().then((d) => {
        if (d != null) {
          durationRef.current = d;
          setDuration(d);
        }
      });
    });

    p.addEventListener("ended", () => {
      endedRef.current = true;
      isPlayingRef.current = false;
      setIsPlaying(false);
      clearTimer();
    });
  };

  const setupPlayer = (startAt = 0) => {
    if (!containerRef.current) return;
    const p = create(
      { data: logs },
      containerRef.current,
      {
        fit: "width",
        controls: false,
        speed: speedRef.current,
        startAt,
      },
    );
    playerRef.current = p;
    attachListeners(p);
    p.play();
  };

  // Initial setup — runs once
  useEffect(() => {
    setupPlayer();
    return () => {
      clearTimer();
      playerRef.current?.dispose();
      playerRef.current = null;
    };
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const seekTo = (t: number) => {
    const wasPlaying = isPlayingRef.current;
    const clamped = Math.max(0, Math.min(durationRef.current, t));
    playerRef.current?.pause();
    void playerRef.current?.seek(clamped).then(() => {
      if (wasPlaying) playerRef.current?.play();
    });
    currentTimeRef.current = clamped;
    setCurrentTime(clamped);
  };

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLSelectElement) return;

      switch (e.key) {
        case " ":
          e.preventDefault();
          if (isPlayingRef.current) {
            playerRef.current?.pause();
            isPlayingRef.current = false;
            setIsPlaying(false);
            clearTimer();
          } else {
            playerRef.current?.play();
            isPlayingRef.current = true;
            setIsPlaying(true);
          }
          break;
        case "ArrowLeft":
          e.preventDefault();
          seekTo(currentTimeRef.current - (e.shiftKey ? durationRef.current * 0.1 : 5));
          break;
        case "ArrowRight":
          e.preventDefault();
          seekTo(currentTimeRef.current + (e.shiftKey ? durationRef.current * 0.1 : 5));
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
          if (document.fullscreenElement) {
            void document.exitFullscreen();
          } else {
            onClose?.();
          }
          break;
        default:
          if (e.key >= "0" && e.key <= "9" && !e.shiftKey && !e.ctrlKey && !e.metaKey) {
            seekTo(durationRef.current * parseInt(e.key) / 10);
          }
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [onClose]);

  const handlePlayPause = () => {
    if (isPlayingRef.current) {
      playerRef.current?.pause();
      isPlayingRef.current = false;
      setIsPlaying(false);
      clearTimer();
    } else {
      playerRef.current?.play();
      isPlayingRef.current = true;
      setIsPlaying(true);
    }
  };

  const handleSeek = (value: number) => {
    const wasPlaying = isPlayingRef.current;
    playerRef.current?.pause();
    void playerRef.current?.seek(value).then(() => {
      if (wasPlaying) playerRef.current?.play();
    });
    currentTimeRef.current = value;
    setCurrentTime(value);
  };

  const handleSpeedChange = (newSpeed: Speed) => {
    speedRef.current = newSpeed;
    setSpeed(newSpeed);
    const startAt = endedRef.current ? 0 : currentTimeRef.current;
    playerRef.current?.dispose();
    clearTimer();
    setupPlayer(startAt);
  };

  return (
    <div ref={rootRef} className="flex flex-col h-full bg-[#121314]">
      {/* Player container */}
      <div className="flex-1 min-h-0 overflow-y-auto">
        <div ref={containerRef} className="w-full" />
      </div>

      {/* Controls */}
      <div className="relative flex items-center gap-3 px-4 py-3 bg-surface border-t border-border shrink-0">
        <button
          onClick={handlePlayPause}
          className="w-8 h-8 rounded-full bg-primary hover:bg-primary/90 flex items-center justify-center shrink-0 transition-colors"
          aria-label={isPlaying ? "Pause" : "Play"}
        >
          {isPlaying ? (
            <PauseIcon className="w-3.5 h-3.5 text-white" />
          ) : (
            <PlayIcon className="w-3.5 h-3.5 text-white ml-0.5" />
          )}
        </button>

        <span className="text-xs font-mono tabular-nums text-text-secondary shrink-0">
          {formatTime(currentTime)} / {formatTime(duration)}
        </span>

        <input
          type="range"
          min={0}
          max={duration || 1}
          step={0.1}
          value={currentTime}
          onChange={(e) => handleSeek(parseFloat(e.target.value))}
          className="flex-1 accent-primary"
          aria-label="Seek"
        />

        <select
          value={speed}
          onChange={(e) => handleSpeedChange(parseFloat(e.target.value) as Speed)}
          className="bg-surface border border-border rounded text-xs font-mono text-text-secondary px-2 py-1 shrink-0"
          aria-label="Playback speed"
        >
          <option value={0.5}>0.5x</option>
          <option value={1}>1x</option>
          <option value={1.5}>1.5x</option>
          <option value={2}>2x</option>
        </select>

        <button
          onClick={() => setShowShortcuts((v) => !v)}
          className={`w-7 h-7 flex items-center justify-center rounded-md transition-colors shrink-0 ${
            showShortcuts
              ? "bg-primary/10 text-primary border border-primary/20"
              : "text-text-muted hover:text-text-primary hover:bg-card border border-transparent"
          }`}
          aria-label="Keyboard shortcuts"
          title="Keyboard shortcuts"
        >
          <KeyboardIcon className="w-4 h-4" />
        </button>

        {/* Shortcuts popover */}
        {showShortcuts && (
          <div className="absolute bottom-full right-4 mb-2 bg-card border border-border rounded-lg shadow-lg p-3 w-80">
            <p className="text-2xs font-mono font-semibold uppercase tracking-widest text-text-muted/60 mb-2.5">
              Keyboard Shortcuts
            </p>
            <div className="space-y-1.5">
              {SHORTCUTS.map(({ keys, description }) => (
                <div key={description} className="flex items-center justify-between gap-4">
                  <span className="text-xs text-text-secondary">{description}</span>
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
          </div>
        )}
      </div>
    </div>
  );
}
