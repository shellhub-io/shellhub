import { create } from "zustand";
import { listRecordings, type RecordingMeta } from "@/utils/recordings";

interface RecordingsState {
  recordings: RecordingMeta[];
  notice: RecordingMeta | null;
  refresh: () => Promise<void>;
  notify: (meta: RecordingMeta) => void;
  clearNotice: () => void;
}

/**
 * The locally recorded sessions, listed from OPFS.
 */
export const useRecordingsStore = create<RecordingsState>((set, get) => ({
  recordings: [],
  notice: null,

  refresh: async () => {
    set({ recordings: await listRecordings() });
  },

  notify: (meta) => {
    set({ notice: meta });
    void get().refresh();
  },

  clearNotice: () => set({ notice: null }),
}));
