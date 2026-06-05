import { useEffect, useRef, useState } from "react";
import { useParams } from "react-router-dom";
import { Terminal } from "@xterm/xterm";
import { WebglAddon } from "@xterm/addon-webgl";
import { WebLinksAddon } from "@xterm/addon-web-links";
import { CommandLineIcon, Cog6ToothIcon } from "@heroicons/react/24/outline";
import { useTerminalThemeStore } from "@/stores/terminalThemeStore";
import TerminalSettingsDrawer from "@/components/terminal/TerminalSettingsDrawer";

type Status = "connecting" | "watching" | "ended" | "invalid";

// SharedTerminal is a public, read-only viewer for a terminal session shared by an agent
// (tmate/upterm style). It requires no authentication: the unguessable token in the URL is the
// only capability. The viewer mirrors the host's geometry and never sends input.
export default function SharedTerminal() {
  const { token } = useParams<{ token: string }>();
  const containerRef = useRef<HTMLDivElement>(null);
  const termRef = useRef<Terminal | null>(null);
  const writableRef = useRef(false);
  const [status, setStatus] = useState<Status>("connecting");
  const [writable, setWritable] = useState(false);
  const [name, setName] = useState("");
  const [settingsOpen, setSettingsOpen] = useState(false);

  const { theme, fontFamilyWithFallback, fontSize, loadThemes } =
    useTerminalThemeStore();

  useEffect(() => {
    document.title = "ShellHub — Shared terminal";
    void loadThemes();
  }, [loadThemes]);

  useEffect(() => {
    if (!token) return;

    const {
      theme: initTheme,
      fontFamilyWithFallback: initFont,
      fontSize: initSize,
    } = useTerminalThemeStore.getState();

    const term = new Terminal({
      theme: initTheme.colors,
      fontFamily: initFont,
      fontSize: initSize,
      cursorBlink: false,
      disableStdin: true,
      allowProposedApi: true,
    });
    termRef.current = term;

    term.loadAddon(new WebLinksAddon());

    // The viewer mirrors the host's exact geometry (sent as resize control frames), so we don't
    // fit to the container — the fixed-size grid is centered and letterboxed instead, like tmate.
    if (containerRef.current) {
      term.open(containerRef.current);
      try {
        term.loadAddon(new WebglAddon());
      } catch {
        // DOM renderer fallback
      }
    }

    const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
    const ws = new WebSocket(`${proto}//${window.location.host}/ws/share/${token}`);
    let opened = false;

    ws.onopen = () => {
      opened = true;
      setStatus("watching");
    };

    // In collaborative mode, forward the viewer's keystrokes to the host as binary frames.
    term.onData((data) => {
      if (writableRef.current && ws.readyState === WebSocket.OPEN) {
        ws.send(new TextEncoder().encode(data));
      }
    });

    ws.onmessage = async (event) => {
      if (event.data instanceof Blob) {
        // Binary frame = raw terminal output.
        term.write(new Uint8Array(await event.data.arrayBuffer()));
        return;
      }

      // Text frame = JSON control message (init handshake or resize).
      try {
        const msg = JSON.parse(String(event.data)) as {
          kind?: string;
          cols?: number;
          rows?: number;
          writable?: boolean;
          name?: string;
        };
        if (msg.kind === "init") {
          const w = Boolean(msg.writable);
          writableRef.current = w;
          setWritable(w);
          term.options.disableStdin = !w;
          term.options.cursorBlink = w;
          if (w) term.focus();
          if (msg.name) {
            setName(msg.name);
            document.title = `ShellHub — ${msg.name}`;
          }
        } else if (msg.kind === "resize" && msg.cols && msg.rows) {
          term.resize(msg.cols, msg.rows);
        }
      } catch {
        // Ignore malformed control frames.
      }
    };

    // A socket that closes without ever opening means the link is invalid or expired; one that
    // opened and then closed means the host ended the session.
    ws.onclose = () => setStatus(opened ? "ended" : "invalid");
    ws.onerror = () => setStatus(opened ? "ended" : "invalid");

    return () => {
      ws.onopen = null;
      ws.onmessage = null;
      ws.onclose = null;
      ws.onerror = null;
      ws.close();
      term.dispose();
      termRef.current = null;
    };
  }, [token]);

  // Live theme/font updates, mirroring the authenticated terminal.
  useEffect(() => {
    if (termRef.current) termRef.current.options.theme = theme.colors;
  }, [theme]);

  useEffect(() => {
    if (termRef.current) termRef.current.options.fontFamily = fontFamilyWithFallback;
  }, [fontFamilyWithFallback]);

  useEffect(() => {
    if (termRef.current) termRef.current.options.fontSize = fontSize;
  }, [fontSize]);

  const dot =
    status === "watching"
      ? "bg-accent-green shadow-[0_0_6px_rgba(130,165,104,0.4)]"
      : status === "connecting"
        ? "bg-accent-yellow animate-pulse-subtle"
        : "bg-text-muted/50";

  const label =
    status === "watching"
      ? writable
        ? "Live (you can type)"
        : "Live (read-only)"
      : status === "connecting"
        ? "Connecting…"
        : status === "invalid"
          ? "Link invalid or expired"
          : "Session ended";

  return (
    <div className="flex h-screen flex-col bg-background text-text-primary">
      <header className="flex items-center justify-between border-b border-border px-6 py-3">
        <div className="flex items-baseline gap-2">
          <span className="text-sm font-semibold text-text-primary">ShellHub</span>
          <span className="text-text-muted/40">/</span>
          <span className="text-sm text-text-secondary">
            {name || "Shared terminal"}
          </span>
        </div>
        <div className="flex items-center gap-4">
          {writable && status === "watching" && (
            <span className="rounded border border-primary/30 bg-primary/10 px-2 py-0.5 text-2xs font-semibold uppercase tracking-label text-primary">
              Collaborative
            </span>
          )}
          <span className="flex items-center gap-2 text-xs">
            <span className={`h-2 w-2 rounded-full ${dot}`} />
            <span className="text-text-secondary">{label}</span>
          </span>
          <button
            onClick={() => setSettingsOpen(true)}
            title="Terminal settings"
            className="rounded-md p-1.5 text-text-muted transition-colors hover:bg-surface hover:text-text-primary"
          >
            <Cog6ToothIcon className="h-4 w-4" />
          </button>
        </div>
      </header>

      <div className="relative flex flex-1 items-center justify-center overflow-auto p-6">
        {/* Framed terminal window so the host-sized grid reads as a deliberate panel rather than
            floating loose in the viewport. */}
        <div
          className="rounded-lg border border-border-light p-3 shadow-2xl shadow-black/40"
          style={{ backgroundColor: theme.colors.background }}
        >
          <div ref={containerRef} />
        </div>

        {(status === "ended" || status === "invalid") && (
          <div className="absolute inset-0 flex items-center justify-center bg-background/80 backdrop-blur-sm">
            <div className="text-center">
              <CommandLineIcon
                className="mx-auto mb-3 h-10 w-10 text-text-muted/30"
                strokeWidth={1}
              />
              <p className="text-sm text-text-secondary">
                {status === "invalid"
                  ? "This share link is invalid or has expired."
                  : "This shared terminal has ended."}
              </p>
              <p className="mt-1 font-mono text-xs text-text-muted">
                {status === "invalid"
                  ? "Ask for a fresh link to watch."
                  : "The host closed the session."}
              </p>
            </div>
          </div>
        )}
      </div>

      <TerminalSettingsDrawer
        open={settingsOpen}
        onClose={() => setSettingsOpen(false)}
      />
    </div>
  );
}
