import { create } from "zustand";
import { getSessions, getSession, closeSession } from "../api/sessions";
import { Session } from "../types/session";

interface SessionsState {
  sessions: Session[];
  session: Session | null;
  totalCount: number;
  loading: boolean;
  error: string | null;
  page: number;
  perPage: number;
  fetch: (page?: number, perPage?: number) => Promise<void>;
  fetchOne: (uid: string) => Promise<void>;
  close: (uid: string) => Promise<void>;
  setPage: (page: number) => void;
}

export const useSessionsStore = create<SessionsState>((set, get) => ({
  sessions: [],
  session: null,
  totalCount: 0,
  loading: false,
  error: null,
  page: 1,
  perPage: 10,

  fetch: async (page?: number, perPage?: number) => {
    const p = page ?? get().page;
    const pp = perPage ?? get().perPage;
    set({ loading: true, error: null });
    try {
      const { data, totalCount } = await getSessions(p, pp);
      set({ sessions: data, totalCount, loading: false, page: p, perPage: pp });
    } catch {
      set({ loading: false, error: "Failed to load sessions" });
    }
  },

  fetchOne: async (uid: string) => {
    set({ loading: true, error: null, session: null });
    try {
      const data = await getSession(uid);
      set({ session: data, loading: false });
    } catch {
      set({ loading: false, error: "Failed to load session" });
    }
  },

  close: async (uid: string) => {
    await closeSession(uid);
    const current = get().session;
    if (current?.uid === uid) {
      set({ session: { ...current, active: false } });
    }
    set((s) => ({
      sessions: s.sessions.map((sess) =>
        sess.uid === uid ? { ...sess, active: false } : sess,
      ),
    }));
  },

  setPage: (page: number) => set({ page }),
}));
