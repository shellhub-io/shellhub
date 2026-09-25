import { create } from "zustand";
import { generateRandomUUID } from "@/utils/random-uuid";
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

interface TerminalState {
  sessions: TerminalSession[];
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
  closeAndReconnect: (id: string) => void;
  requestConnect: (deviceUid: string, deviceName: string) => void;
  clearReconnect: () => void;
  setConnectionStatus: (id: string, status: ConnectionStatus) => void;
  clearSensitiveData: (id: string) => void;
}

function demoteOthers(
  sessions: TerminalSession[],
  targetId: string,
): TerminalSession[] {
  return sessions.map((s) => {
    if (s.id === targetId) return s;
    if (s.state !== "minimized") return { ...s, state: "minimized" as const };
    return s;
  });
}

/**
 * The open terminals. Several may run at once, so this is a list rather than one session, and
 * opening one also records the device as recently used.
 */
export const useTerminalStore = create<TerminalState>((set, get) => ({
  sessions: [],
  reconnectTarget: null,
  restoreAfterNavigation: null,

  setRestoreAfterNavigation: (id) => set({ restoreAfterNavigation: id }),

  dockPendingRestore: () => {
    const pending = get().restoreAfterNavigation;
    if (!pending) return false;
    set((state) => ({
      restoreAfterNavigation: null,
      sessions: demoteOthers(state.sessions, pending).map((s) =>
        s.id === pending ? { ...s, state: "docked" as const } : s,
      ),
    }));
    return true;
  },

  open: (params) => {
    const id = generateRandomUUID();
    useRecentDevicesStore
      .getState()
      .record(params.deviceUid, params.deviceName);
    set((state) => ({
      reconnectTarget: null,
      sessions: [
        ...demoteOthers(state.sessions, id),
        { ...params, id, state: "docked", connectionStatus: "connecting" },
      ],
    }));
  },

  minimize: (id) => {
    set((state) => ({
      sessions: state.sessions.map((s) =>
        s.id === id ? { ...s, state: "minimized" as const } : s,
      ),
    }));
  },

  minimizeAll: () => {
    set((state) => ({
      sessions: state.sessions.map((s) =>
        s.state !== "minimized" ? { ...s, state: "minimized" as const } : s,
      ),
    }));
  },

  restore: (id) => {
    set((state) => ({
      sessions: demoteOthers(state.sessions, id).map((s) =>
        s.id === id ? { ...s, state: "docked" as const } : s,
      ),
    }));
  },

  toggleFullscreen: (id) => {
    set((state) => ({
      sessions: demoteOthers(state.sessions, id).map((s) => {
        if (s.id !== id) return s;
        return {
          ...s,
          state:
            s.state === "fullscreen"
              ? ("docked" as const)
              : ("fullscreen" as const),
        };
      }),
    }));
  },

  close: (id) => {
    set((state) => ({
      sessions: state.sessions.filter((s) => s.id !== id),
    }));
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
