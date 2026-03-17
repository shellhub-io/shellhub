import { create } from "zustand";
import { getInfo } from "../client";

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
      await getInfo({ throwOnError: true });
      useConnectivityStore.getState().markUp();
      polling = false;
    } catch {
      setTimeout(() => void poll(), 5000);
    }
  };

  setTimeout(() => void poll(), 5000);
}

export const useConnectivityStore = create<ConnectivityState>()((set) => ({
  apiReachable: true,
  initialCheckDone: false,
  initialGatePassed: false,

  checkInitial: async () => {
    try {
      await getInfo({ throwOnError: true });
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
    set({ apiReachable: true, initialGatePassed: true });
  },
}));
