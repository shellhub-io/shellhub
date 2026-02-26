import { useEffect, useRef, useCallback, useState } from "react";
import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import { WebLinksAddon } from "@xterm/addon-web-links";
import { WebglAddon } from "@xterm/addon-webgl";
import apiClient from "../../api/client";
import type { TerminalSession } from "../../stores/terminalStore";
import { useTerminalStore } from "../../stores/terminalStore";
import { useTerminalThemeStore } from "../../stores/terminalThemeStore";
import type { TerminalError } from "./terminalErrors";
import TerminalErrorBanner from "./TerminalErrorBanner";
import {
  WS_KIND,
  HTTP_CONNECT_ERROR,
  WS_CLOSE_ERROR,
  WS_NETWORK_ERROR,
  parseMessage,
  resolveError,
} from "./terminalErrors";

interface TerminalInstanceProps {
  session: TerminalSession;
  visible: boolean;
}

export default function TerminalInstance({
  session,
  visible,
}: TerminalInstanceProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const termRef = useRef<Terminal | null>(null);
  const fitRef = useRef<FitAddon | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const observerRef = useRef<ResizeObserver | null>(null);
  const prevVisibleRef = useRef(visible);
  const [error, setError] = useState<TerminalError | null>(null);

  const { theme, fontFamilyWithFallback, fontSize } = useTerminalThemeStore();

  const updateStatus = useCallback(
    (s: "connecting" | "connected" | "disconnected") => {
      useTerminalStore.getState().setConnectionStatus(session.id, s);
    },
    [session.id],
  );

  // Connect on mount, cleanup on unmount
  useEffect(() => {
    let cancelled = false;
    let lastError = false;
    const {
      theme: initTheme,
      fontFamilyWithFallback: initFont,
      fontSize: initSize,
    } = useTerminalThemeStore.getState();

    async function connect() {
      updateStatus("connecting");

      let token: string;
      try {
        const res = await apiClient.post("/ws/ssh", {
          device: session.deviceUid,
          username: session.username,
          password: session.password,
        });
        token = res.data.token;
      } catch {
        if (cancelled) return;
        updateStatus("disconnected");
        setError(HTTP_CONNECT_ERROR);
        return;
      }

      if (cancelled) return;

      const term = new Terminal({
        theme: initTheme.colors,
        fontFamily: initFont,
        fontSize: initSize,
        cursorBlink: true,
        allowProposedApi: true,
      });
      termRef.current = term;

      const fitAddon = new FitAddon();
      fitRef.current = fitAddon;
      term.loadAddon(fitAddon);
      term.loadAddon(new WebLinksAddon());

      if (!containerRef.current) return;
      term.open(containerRef.current);

      try {
        term.loadAddon(new WebglAddon());
      } catch {
        // DOM renderer fallback
      }

      fitAddon.fit();
      const { cols, rows } = term;

      const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
      const wsUrl = `${proto}//${window.location.host}/ws/ssh?token=${token}&cols=${cols}&rows=${rows}`;
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        if (cancelled) return;
        updateStatus("connected");
      };

      ws.onmessage = async (event) => {
        if (cancelled) return;
        if (event.data instanceof Blob) {
          term.write(await event.data.text());
        } else {
          const msg = parseMessage(event.data);
          if (msg?.kind === WS_KIND.ERROR) {
            lastError = true;
            updateStatus("disconnected");
            setError(resolveError(msg.data, session.deviceUid));
          } else if (!lastError) {
            term.write(event.data);
          }
        }
      };

      ws.onclose = () => {
        if (cancelled) return;
        updateStatus("disconnected");
        if (!lastError) {
          setError(WS_CLOSE_ERROR);
        }
      };

      ws.onerror = () => {
        if (cancelled) return;
        lastError = true;
        updateStatus("disconnected");
        setError(WS_NETWORK_ERROR);
      };

      term.onData((data) => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(
            JSON.stringify({ kind: WS_KIND.INPUT, data: data.slice(0, 4096) }),
          );
        }
      });

      term.onResize(({ cols, rows }) => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(
            JSON.stringify({ kind: WS_KIND.RESIZE, data: { cols, rows } }),
          );
        }
      });

      const observer = new ResizeObserver(() => {
        if (fitRef.current) fitAddon.fit();
      });
      observer.observe(containerRef.current);
      observerRef.current = observer;

      term.focus();
    }

    connect();

    return () => {
      cancelled = true;
      observerRef.current?.disconnect();
      observerRef.current = null;
      if (wsRef.current) {
        wsRef.current.onopen = null;
        wsRef.current.onclose = null;
        wsRef.current.onerror = null;
        wsRef.current.onmessage = null;
        wsRef.current.close();
        wsRef.current = null;
      }
      termRef.current?.dispose();
      termRef.current = null;
      fitRef.current = null;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [session.id]);

  // Live theme/font updates
  useEffect(() => {
    const term = termRef.current;
    if (!term) return;
    term.options.theme = theme.colors;
  }, [theme]);

  useEffect(() => {
    const term = termRef.current;
    if (!term) return;
    term.options.fontFamily = fontFamilyWithFallback;
    fitRef.current?.fit();
  }, [fontFamilyWithFallback]);

  useEffect(() => {
    const term = termRef.current;
    if (!term) return;
    term.options.fontSize = fontSize;
    fitRef.current?.fit();
  }, [fontSize]);

  // Hide cursor on error
  useEffect(() => {
    const term = termRef.current;
    if (!term || error === null) return;
    term.options.cursorBlink = false;
    term.blur();
    term.write("\x1b[?25l"); // CSI sequence to hide cursor
  }, [error]);

  // Handle visibility changes (minimize/restore)
  useEffect(() => {
    if (!prevVisibleRef.current && visible && error === null) {
      requestAnimationFrame(() => {
        fitRef.current?.fit();
        termRef.current?.focus();
      });
    }
    prevVisibleRef.current = visible;
  }, [visible, error]);

  return (
    <div className="relative flex-1 flex flex-col overflow-hidden">
      {error !== null && (
        <TerminalErrorBanner error={error} sessionId={session.id} />
      )}
      <div
        ref={containerRef}
        className={`flex-1 ${error !== null ? "opacity-30 pointer-events-none" : ""}`}
      />
    </div>
  );
}
