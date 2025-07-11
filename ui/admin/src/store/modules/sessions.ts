import { defineStore } from "pinia";
import { IAdminSession } from "@admin/interfaces/ISession";
import * as apiSession from "../api/sessions";

export const useSessionsStore = defineStore("sessions", {
  state: () => ({
    sessions: [] as Array<IAdminSession>,
    session: {} as IAdminSession,
    numberSessions: 0,
  }),

  getters: {
    getSessions: (state) => state.sessions,
    getSession: (state) => state.session,
    getNumberSessions: (state) => state.numberSessions,
  },

  actions: {
    async fetch(data: { perPage: number; page: number }) {
      const res = await apiSession.fetchSessions(data.perPage, data.page);

      if (res.data.length) {
        this.sessions = res.data as Array<IAdminSession>;
        this.numberSessions = parseInt(res.headers["x-total-count"], 10);
        return res;
      }

      return false;
    },

    async get(uid: string) {
      const res = await apiSession.getSession(uid);
      this.session = res.data as IAdminSession;
    },

    clearListSessions() {
      this.sessions = [];
      this.numberSessions = 0;
    },

    clearObjectSession() {
      this.session = {} as IAdminSession;
    },
  },
});

export default useSessionsStore;
