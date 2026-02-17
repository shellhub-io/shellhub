import { create } from "zustand";
import apiClient from "../api/client";

interface ConnectivityState {
  apiReachable: boolean;
  initialCheckDone: boolean;
  /** Once true, the startup gate is open forever — banner handles the rest. */
  initialGatePassed: boolean;
  checkInitial: () => Promise<void>;
  markDown: () => void;
  markUp: () => void;
}

let polling = false;

function startPolling() {
  if (polling) return;
  polling = true;

  const poll = async () => {
    try {
      await apiClient.get("/info", { timeout: 5000 });
      useConnectivityStore.getState().markUp();
      polling = false;
    } catch {
      setTimeout(poll, 5000);
    }
  };

  setTimeout(poll, 5000);
}

export const useConnectivityStore = create<ConnectivityState>()((set) => ({
  apiReachable: true,
  initialCheckDone: false,
  initialGatePassed: false,

  checkInitial: async () => {
    try {
      await apiClient.get("/info", { timeout: 5000 });
      set({
        apiReachable: true,
        initialCheckDone: true,
        initialGatePassed: true,
      });
    } catch {
      set({ apiReachable: false, initialCheckDone: true });
      startPolling();
    }
  },

  markDown: () => {
    set({ apiReachable: false });
    startPolling();
  },

  markUp: () => {
    set((s) => ({
      apiReachable: true,
      initialGatePassed: s.initialGatePassed || true,
    }));
  },
}));
