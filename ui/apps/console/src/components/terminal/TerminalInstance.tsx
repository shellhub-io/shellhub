import { useEffect, useRef, useCallback, useState } from "react";
import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import { WebLinksAddon } from "@xterm/addon-web-links";
import { WebglAddon } from "@xterm/addon-webgl";
import { Buffer } from "buffer";
import { createWebSshSession, type CreateWebSshSessionData } from "@/client";
import { generateSignature } from "@/utils/sshKeys";
import type { TerminalSession } from "@/stores/terminalStore";
import { useTerminalStore } from "@/stores/terminalStore";
import { useTerminalThemeStore } from "@/stores/terminalThemeStore";
import { useAuthStore } from "@/stores/authStore";
import { useNamespace } from "@/hooks/useNamespaces";
import { useRecordingsStore } from "@/stores/recordingsStore";
import { OpfsCastRecorder } from "@/utils/recordings";
import { nextFontSize } from "./fontSizeShortcut";
import type { TerminalError } from "./terminalErrors";
import TerminalErrorBanner from "./TerminalErrorBanner";
import {
  WS_KIND,
  HTTP_CONNECT_ERROR,
  WS_CLOSE_ERROR,
  WS_NETWORK_ERROR,
  WS_REAUTH_CANCELLED,
  parseMessage,
  resolveError,
} from "./terminalErrors";
import SSHApproval from "@/pages/SSHApproval";
import { cn } from "@shellhub/design-system/cn";
import { attempt } from "@/utils/failure";

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
  const resizeRegisteredRef = useRef(false);
  const recorderRef = useRef<OpfsCastRecorder | null>(null);
  const [error, setError] = useState<TerminalError | null>(null);
  const [approvalCode, setApprovalCode] = useState<string | null>(null);

  const { theme, fontFamilyWithFallback, fontSize } = useTerminalThemeStore();

  const tenant = useAuthStore((s) => s.tenant);
  const { namespace } = useNamespace(tenant ?? "");
  const isIdentityMode = namespace?.settings?.ssh_access_mode === "identity";

  const updateStatus = useCallback(
    (s: "connecting" | "connected" | "disconnected") => {
      useTerminalStore.getState().setConnectionStatus(session.id, s);
    },
    [session.id],
  );

  useEffect(() => {
    let cancelled = false;
    let lastError = false;
    const {
      theme: initTheme,
      fontFamilyWithFallback: initFont,
      fontSize: initSize,
    } = useTerminalThemeStore.getState();

    async function finalizeRecording() {
      const recorder = recorderRef.current;
      if (!recorder) return;
      recorderRef.current = null;
      const meta = await recorder.finish();
      if (meta) useRecordingsStore.getState().notify(meta);
    }

    async function connect() {
      updateStatus("connecting");

      const shared = { device: session.deviceUid, username: session.username };
      const body: CreateWebSshSessionData["body"] = session.fingerprint
        ? {
            ...shared,
            fingerprint: session.fingerprint,
            public_key: session.publicKeyLine,
          }
        : { ...shared, password: session.password ?? "" };

      let token: string;
      try {
        const { data } = await createWebSshSession({
          body,
          throwOnError: true,
        });
        token = data.token;
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

      term.attachCustomKeyEventHandler((event) => {
        if (event.type !== "keydown") return true;

        const { fontSize, setFontSize } = useTerminalThemeStore.getState();
        const next = nextFontSize(fontSize, event);
        if (next === null) return true;

        setFontSize(next);

        event.preventDefault();

        return false;
      });

      const fitAddon = new FitAddon();
      fitRef.current = fitAddon;
      term.loadAddon(fitAddon);
      term.loadAddon(new WebLinksAddon());

      if (!containerRef.current) return;
      term.open(containerRef.current);

      attempt(() => term.loadAddon(new WebglAddon()));

      fitAddon.fit();
      const { cols, rows } = term;

      const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
      const wsUrl = `${proto}//${window.location.host}/ws/ssh?token=${token}&cols=${cols}&rows=${rows}`;
      const ws = new WebSocket(wsUrl);
      ws.binaryType = "arraybuffer";
      wsRef.current = ws;
      resizeRegisteredRef.current = false;

      const recordingDecoder = new TextDecoder("utf-8");

      let keyMaterial: string | undefined = session.privateKey;
      let keyPassphrase: string | undefined = session.passphrase;
      const browserKey = session.browserKey;

      const registerResizeHandler = () => {
        if (resizeRegisteredRef.current) return;
        resizeRegisteredRef.current = true;

        term.onResize(({ cols, rows }) => {
          if (ws.readyState === WebSocket.OPEN) {
            ws.send(
              JSON.stringify({ kind: WS_KIND.RESIZE, data: { cols, rows } }),
            );
          }
          recorderRef.current?.recordResize(cols, rows);
        });
      };

      ws.onopen = async () => {
        if (cancelled) return;
        updateStatus("connected");

        if (session.record) {
          try {
            const recorder = await OpfsCastRecorder.create(
              session.deviceName,
              session.deviceUid,
              session.username,
            );
            if (cancelled) {
              await recorder.discard();
              return;
            }
            recorderRef.current = recorder;
            recorder.start(cols, rows);
          } catch (err) {
            console.error("session recording: could not start", err);
          }
        }
      };

      ws.onmessage = (event) => {
        if (cancelled) return;
        if (typeof event.data !== "string") {
          const bytes = new Uint8Array(event.data as ArrayBuffer);
          term.write(bytes);
          const decoded = recordingDecoder.decode(bytes, { stream: true });
          recorderRef.current?.recordOutput(decoded);
          registerResizeHandler();
          return;
        }

        void (async () => {
          const textData = String(event.data as unknown);
          const msg = parseMessage(textData);
          if (!msg) {
            if (!lastError) term.write(textData);
            return;
          }

          switch (msg.kind) {
            case WS_KIND.SIGNATURE: {
              if (!browserKey && !keyMaterial) return;
              const challengeBuffer = Buffer.from(msg.data, "base64");
              try {
                let signature: string;
                if (browserKey) {
                  const sig = await crypto.subtle.sign(
                    { name: "Ed25519" },
                    browserKey,
                    challengeBuffer,
                  );
                  signature = Buffer.from(sig).toString("base64");
                } else {
                  signature = generateSignature(
                    keyMaterial as string,
                    challengeBuffer,
                    keyPassphrase,
                  );
                }
                ws.send(
                  JSON.stringify({ kind: WS_KIND.SIGNATURE, data: signature }),
                );
              } catch {
                term.write(
                  "\r\n\x1b[1;31mFailed to sign authentication challenge.\x1b[0m\r\n",
                );
                keyMaterial = undefined;
                keyPassphrase = undefined;
                useTerminalStore.getState().clearSensitiveData(session.id);
                ws.close();
                return;
              }
              keyMaterial = undefined;
              keyPassphrase = undefined;
              useTerminalStore.getState().clearSensitiveData(session.id);
              registerResizeHandler();
              break;
            }
            case WS_KIND.ERROR: {
              lastError = true;
              updateStatus("disconnected");
              setError(resolveError(msg.data, session.deviceUid, isIdentityMode));
              break;
            }
            case WS_KIND.SESSION: {
              recorderRef.current?.setSessionUid(msg.data);
              break;
            }
            case WS_KIND.REAUTH: {
              setApprovalCode(msg.data);
              break;
            }
            default:
              break;
          }
        })();
      };

      ws.onclose = () => {
        if (cancelled) return;
        updateStatus("disconnected");
        if (!lastError) {
          setError(WS_CLOSE_ERROR);
        }
        void finalizeRecording();
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

      const observer = new ResizeObserver(() => {
        if (fitRef.current) fitAddon.fit();
      });
      observer.observe(containerRef.current);
      observerRef.current = observer;

      term.focus();
    }

    void connect();

    return () => {
      cancelled = true;
      const stillOpen = useTerminalStore
        .getState()
        .sessions.some((s) => s.id === session.id);
      if (stillOpen) {
        const recorder = recorderRef.current;
        recorderRef.current = null;
        void recorder?.discard();
      } else {
        void finalizeRecording();
      }
      useTerminalStore.getState().clearSensitiveData(session.id);
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

  useEffect(() => {
    const term = termRef.current;
    if (!term || error === null) return;
    term.options.cursorBlink = false;
    term.blur();
    term.write("\x1b[?25l"); // CSI sequence to hide cursor
  }, [error]);

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
      {session.record && session.connectionStatus === "connected" && (
        <div className="absolute top-2 right-3 z-10 flex items-center gap-1.5 px-2 py-0.5 rounded-md bg-black/40 backdrop-blur-sm pointer-events-none select-none">
          <span className="w-1.5 h-1.5 rounded-full bg-accent-red animate-pulse shadow-[0_0_4px_rgba(220,80,80,0.7)]" />
          <span className="text-2xs font-semibold tracking-wide text-accent-red">
            REC
          </span>
        </div>
      )}
      <div
        ref={containerRef}
        className={cn(
          "flex-1",
          error !== null && "opacity-30 pointer-events-none",
        )}
      />
      {approvalCode && (
        <SSHApproval
          flow="confirm"
          code={approvalCode}
          onClose={() => {
            setApprovalCode(null);
            setError(WS_REAUTH_CANCELLED);
          }}
          onDecided={() => {
            setApprovalCode(null);
            requestAnimationFrame(() => termRef.current?.focus());
          }}
        />
      )}
    </div>
  );
}
