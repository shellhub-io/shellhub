import { create } from "zustand";

export type TerminalWindowState = "docked" | "minimized" | "fullscreen";
export type ConnectionStatus = "connecting" | "connected" | "disconnected";

export interface TerminalSession {
  id: string;
  deviceUid: string;
  deviceName: string;
  username: string;
  password: string;
  state: TerminalWindowState;
  connectionStatus: ConnectionStatus;
}

interface TerminalState {
  sessions: TerminalSession[];
  open: (params: Omit<TerminalSession, "id" | "state" | "connectionStatus">) => void;
  minimize: (id: string) => void;
  minimizeAll: () => void;
  restore: (id: string) => void;
  toggleFullscreen: (id: string) => void;
  close: (id: string) => void;
  setConnectionStatus: (id: string, status: ConnectionStatus) => void;
}

function demoteOthers(sessions: TerminalSession[], targetId: string): TerminalSession[] {
  return sessions.map((s) => {
    if (s.id === targetId) return s;
    if (s.state !== "minimized") return { ...s, state: "minimized" as const };
    return s;
  });
}

export const useTerminalStore = create<TerminalState>((set) => ({
  sessions: [],

  open: (params) => {
    const id = crypto.randomUUID();
    set((state) => ({
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
        return { ...s, state: s.state === "fullscreen" ? "docked" as const : "fullscreen" as const };
      }),
    }));
  },

  close: (id) => {
    set((state) => ({
      sessions: state.sessions.filter((s) => s.id !== id),
    }));
  },

  setConnectionStatus: (id, status) => {
    set((state) => ({
      sessions: state.sessions.map((s) =>
        s.id === id ? { ...s, connectionStatus: status } : s,
      ),
    }));
  },
}));
