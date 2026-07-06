import { create } from "zustand";
import { generateRandomUUID } from "@/utils/random-uuid";
import { useRecentDevicesStore } from "./recentDevicesStore";

export type TerminalWindowState = "docked" | "minimized" | "fullscreen";
export type ConnectionStatus = "connecting" | "connected" | "disconnected";

export interface TerminalSession {
  id: string;
  deviceUid: string;
  deviceName: string;
  username: string;
  password: string;
  fingerprint?: string;
  privateKey?: string;
  passphrase?: string;
  state: TerminalWindowState;
  connectionStatus: ConnectionStatus;
  /** Opt-in: record this session client-side (to OPFS). */
  record?: boolean;
}

export interface ReconnectTarget {
  deviceUid: string;
  deviceName: string;
}

interface TerminalState {
  sessions: TerminalSession[];
  /**
   * Device whose ConnectDrawer should open — set for a fresh connect
   * (`requestConnect`) or a post-close reconnect (`closeAndReconnect`).
   * Consumed by TerminalManager, which builds the sshid and opens the drawer.
   */
  reconnectTarget: ReconnectTarget | null;
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

export const useTerminalStore = create<TerminalState>((set) => ({
  sessions: [],
  reconnectTarget: null,

  open: (params) => {
    const id = generateRandomUUID();
    // The single app-wide connect choke point: every entry point reaches here,
    // so this is where a device joins the palette's Recent list.
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
