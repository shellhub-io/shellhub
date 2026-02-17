import { useEffect, useRef, useCallback } from "react";
import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import { WebLinksAddon } from "@xterm/addon-web-links";
import { WebglAddon } from "@xterm/addon-webgl";
import apiClient from "../../api/client";
import type { TerminalSession } from "../../stores/terminalStore";
import { useTerminalStore } from "../../stores/terminalStore";
import { useTerminalThemeStore } from "../../stores/terminalThemeStore";

interface TerminalInstanceProps {
  session: TerminalSession;
  visible: boolean;
}

export default function TerminalInstance({ session, visible }: TerminalInstanceProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const termRef = useRef<Terminal | null>(null);
  const fitRef = useRef<FitAddon | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const observerRef = useRef<ResizeObserver | null>(null);
  const prevVisibleRef = useRef(visible);

  const { theme, fontFamilyWithFallback, fontSize } = useTerminalThemeStore();

  const updateStatus = useCallback((s: "connecting" | "connected" | "disconnected") => {
    useTerminalStore.getState().setConnectionStatus(session.id, s);
  }, [session.id]);

  // Connect on mount, cleanup on unmount
  useEffect(() => {
    let cancelled = false;
    const { theme: initTheme, fontFamilyWithFallback: initFont, fontSize: initSize } =
      useTerminalThemeStore.getState();

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
        updateStatus("disconnected");
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
        if (event.data instanceof Blob) {
          term.write(await event.data.text());
        } else {
          term.write(event.data);
        }
      };

      ws.onclose = () => {
        if (cancelled) return;
        updateStatus("disconnected");
        term.write("\r\n\x1b[1;31mDisconnected.\x1b[0m\r\n");
      };

      ws.onerror = () => {
        if (cancelled) return;
        updateStatus("disconnected");
        term.write("\r\n\x1b[1;31mConnection error.\x1b[0m\r\n");
      };

      term.onData((data) => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({ kind: 1, data: data.slice(0, 4096) }));
        }
      });

      term.onResize(({ cols, rows }) => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({ kind: 2, data: { cols, rows } }));
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
        wsRef.current.onclose = null;
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

  // Handle visibility changes (minimize/restore)
  useEffect(() => {
    if (!prevVisibleRef.current && visible) {
      requestAnimationFrame(() => {
        fitRef.current?.fit();
        termRef.current?.focus();
      });
    }
    prevVisibleRef.current = visible;
  }, [visible]);

  return (
    <div className="flex-1 overflow-hidden">
      <div ref={containerRef} className="w-full h-full" />
    </div>
  );
}
