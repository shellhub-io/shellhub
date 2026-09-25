import { create } from "zustand";
import { generateRandomUUID } from "@/utils/random-uuid";
import { moveById } from "@/utils/moveById";
import { useRecentDevicesStore } from "./recentDevicesStore";

/**
 * How a terminal window is displayed. Minimized keeps the session alive — the connection is not
 * torn down until the session is closed.
 */
export type TerminalWindowState = "docked" | "minimized" | "fullscreen";
/**
 * Where a terminal's connection stands.
 */
export type ConnectionStatus = "connecting" | "connected" | "disconnected";

/**
 * One open terminal. The credentials are held for the life of the session so a reconnect does
 * not have to ask again; they are in memory only and go with the tab.
 */
export interface TerminalSession {
  id: string;
  deviceUid: string;
  deviceName: string;
  username: string;
  password: string;
  fingerprint?: string;
  privateKey?: string;
  passphrase?: string;
  browserKey?: CryptoKey;
  publicKeyLine?: string;
  state: TerminalWindowState;
  connectionStatus: ConnectionStatus;
  record?: boolean;
  tenant?: string;
}

/**
 * The device a closed terminal offers to reconnect to, which is all that outlives the session.
 * tenant is the namespace the device belongs to, when that may not be the active one.
 */
export interface ReconnectTarget {
  deviceUid: string;
  deviceName: string;
  tenant?: string;
}

/**
 * An open session recording, played in its own tab beside the terminals. The recording is held
 * whole, since it was fetched once to open the tab; id is the recorded session's uid, so playing
 * the same session again brings its tab forward instead of opening another.
 */
export interface RecordingView {
  id: string;
  title: string;
  tenant?: string;
  logs: string;
  shown: boolean;
}

interface TerminalState {
  sessions: TerminalSession[];
  recordings: RecordingView[];
  reconnectTarget: ReconnectTarget | null;
  restoreAfterNavigation: string | null;
  setRestoreAfterNavigation: (id: string | null) => void;
  dockPendingRestore: () => boolean;
  open: (
    params: Omit<TerminalSession, "id" | "state" | "connectionStatus">,
  ) => void;
  minimize: (id: string) => void;
  minimizeAll: () => void;
  restore: (id: string) => void;
  toggleFullscreen: (id: string) => void;
  close: (id: string) => void;
  move: (id: string, to: number) => void;
  closeAndReconnect: (id: string) => void;
  requestConnect: (deviceUid: string, deviceName: string) => void;
  clearReconnect: () => void;
  setConnectionStatus: (id: string, status: ConnectionStatus) => void;
  clearSensitiveData: (id: string) => void;
  openRecording: (params: Omit<RecordingView, "shown">) => void;
  showRecording: (id: string) => void;
  closeRecording: (id: string) => void;
  moveRecording: (id: string, to: number) => void;
}

function bringForward(
  state: Pick<TerminalState, "sessions" | "recordings">,
  forward: { session?: string; recording?: string },
): Pick<TerminalState, "sessions" | "recordings"> {
  return {
    sessions: state.sessions.map((s) =>
      s.id === forward.session || s.state === "minimized"
        ? s
        : { ...s, state: "minimized" as const },
    ),
    recordings: state.recordings.map((r) =>
      r.shown === (r.id === forward.recording)
        ? r
        : { ...r, shown: r.id === forward.recording },
    ),
  };
}

function withSessionState(
  sessions: TerminalSession[],
  id: string,
  next: (s: TerminalSession) => TerminalWindowState,
): TerminalSession[] {
  return sessions.map((s) => (s.id === id ? { ...s, state: next(s) } : s));
}

/**
 * The open terminals and session recordings, the windows that stack over the page inside the
 * content frame. Several may be open at once, but at most one is in view: every action that brings
 * a terminal or a recording forward, or puts them all away, goes through bringForward, which puts
 * every other one away. Opening a terminal also records the device as recently used.
 */
export const useTerminalStore = create<TerminalState>((set, get) => ({
  sessions: [],
  recordings: [],
  reconnectTarget: null,
  restoreAfterNavigation: null,

  setRestoreAfterNavigation: (id) => set({ restoreAfterNavigation: id }),

  dockPendingRestore: () => {
    const pending = get().restoreAfterNavigation;
    if (!pending) return false;
    set((state) => {
      const shown = bringForward(state, { session: pending });
      return {
        restoreAfterNavigation: null,
        recordings: shown.recordings,
        sessions: withSessionState(shown.sessions, pending, () => "docked"),
      };
    });
    return true;
  },

  open: (params) => {
    const id = generateRandomUUID();
    useRecentDevicesStore
      .getState()
      .record(params.deviceUid, params.deviceName);
    set((state) => {
      const shown = bringForward(state, { session: id });
      return {
        reconnectTarget: null,
        recordings: shown.recordings,
        sessions: [
          ...shown.sessions,
          { ...params, id, state: "docked", connectionStatus: "connecting" },
        ],
      };
    });
  },

  minimize: (id) => {
    set((state) => ({
      sessions: state.sessions.map((s) =>
        s.id === id ? { ...s, state: "minimized" as const } : s,
      ),
    }));
  },

  minimizeAll: () => {
    set((state) => bringForward(state, {}));
  },

  restore: (id) => {
    set((state) => {
      const shown = bringForward(state, { session: id });
      return {
        ...shown,
        sessions: withSessionState(shown.sessions, id, () => "docked"),
      };
    });
  },

  toggleFullscreen: (id) => {
    set((state) => {
      const shown = bringForward(state, { session: id });
      return {
        ...shown,
        sessions: withSessionState(shown.sessions, id, (s) =>
          s.state === "fullscreen" ? "docked" : "fullscreen",
        ),
      };
    });
  },

  close: (id) => {
    set((state) => ({
      sessions: state.sessions.filter((s) => s.id !== id),
    }));
  },

  move: (id, to) => {
    set((state) => {
      const sessions = moveById(state.sessions, id, to);
      return sessions === state.sessions ? state : { sessions };
    });
  },

  closeAndReconnect: (id) => {
    set((state) => {
      const session = state.sessions.find((s) => s.id === id);
      if (!session) return state;
      return {
        sessions: state.sessions.filter((s) => s.id !== id),
        reconnectTarget: {
          deviceUid: session.deviceUid,
          deviceName: session.deviceName,
          tenant: session.tenant,
        },
      };
    });
  },

  requestConnect: (deviceUid, deviceName) => {
    set({ reconnectTarget: { deviceUid, deviceName } });
  },

  clearReconnect: () => {
    set({ reconnectTarget: null });
  },

  setConnectionStatus: (id, status) => {
    set((state) => ({
      sessions: state.sessions.map((s) =>
        s.id === id ? { ...s, connectionStatus: status } : s,
      ),
    }));
  },

  clearSensitiveData: (id) => {
    set((state) => ({
      sessions: state.sessions.map((s) =>
        s.id === id
          ? { ...s, privateKey: undefined, passphrase: undefined, password: "" }
          : s,
      ),
    }));
  },

  openRecording: (params) => {
    set((state) => {
      const opened = state.recordings.some((r) => r.id === params.id)
        ? state.recordings.map((r) =>
            r.id === params.id ? { ...params, shown: r.shown } : r,
          )
        : [...state.recordings, { ...params, shown: false }];
      return bringForward(
        { sessions: state.sessions, recordings: opened },
        { recording: params.id },
      );
    });
  },

  showRecording: (id) => {
    set((state) => bringForward(state, { recording: id }));
  },

  closeRecording: (id) => {
    set((state) => ({
      recordings: state.recordings.filter((r) => r.id !== id),
    }));
  },

  moveRecording: (id, to) => {
    set((state) => {
      const recordings = moveById(state.recordings, id, to);
      return recordings === state.recordings ? state : { recordings };
    });
  },
}));

/**
 * Whether a terminal fills the window. The sidebar folds away and the layout drops the room it
 * keeps for it while this holds, so both read it from here to stay in step.
 */
export function useTerminalFullscreen() {
  return useTerminalStore((s) =>
    s.sessions.some((session) => session.state === "fullscreen"),
  );
}
