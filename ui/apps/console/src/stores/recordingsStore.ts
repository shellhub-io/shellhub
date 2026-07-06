import { create } from "zustand";
import {
  listRecordings,
  deleteRecording,
  downloadRecording,
  clearRecordings,
  pruneRecordings,
  type RecordingMeta,
} from "@/utils/recordings";

/** Local retention is a per-browser preference, so it lives in localStorage. */
const RETENTION_KEY = "shellhub:recordings:retentionDays";

function loadRetention(): number | null {
  const raw = localStorage.getItem(RETENTION_KEY);
  const days = raw ? Number(raw) : NaN;
  return Number.isFinite(days) && days > 0 ? days : null;
}

interface RecordingsState {
  recordings: RecordingMeta[];
  loading: boolean;
  /** Auto-delete recordings older than this many days; null = keep forever. */
  retentionDays: number | null;
  /** A just-finished recording surfaced in the post-session snackbar. */
  notice: RecordingMeta | null;
  refresh: () => Promise<void>;
  remove: (id: string) => Promise<void>;
  clearAll: () => Promise<void>;
  setRetentionDays: (days: number | null) => void;
  download: (meta: RecordingMeta) => Promise<void>;
  /** Called when a recording is finalized: refresh the list + raise a notice. */
  notify: (meta: RecordingMeta) => void;
  clearNotice: () => void;
}

export const useRecordingsStore = create<RecordingsState>((set, get) => ({
  recordings: [],
  loading: false,
  retentionDays: loadRetention(),
  notice: null,

  refresh: async () => {
    set({ loading: true });
    try {
      const { retentionDays } = get();
      if (retentionDays) await pruneRecordings(retentionDays);
      set({ recordings: await listRecordings() });
    } finally {
      set({ loading: false });
    }
  },

  remove: async (id) => {
    await deleteRecording(id);
    set((s) => ({ recordings: s.recordings.filter((r) => r.id !== id) }));
  },

  clearAll: async () => {
    await clearRecordings();
    set({ recordings: [] });
  },

  setRetentionDays: (days) => {
    if (days && days > 0) localStorage.setItem(RETENTION_KEY, String(days));
    else localStorage.removeItem(RETENTION_KEY);
    set({ retentionDays: days && days > 0 ? days : null });
    void get().refresh();
  },

  download: async (meta) => {
    await downloadRecording(meta);
  },

  notify: (meta) => {
    set({ notice: meta });
    void get().refresh();
  },

  clearNotice: () => set({ notice: null }),
}));
