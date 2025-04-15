import { defineStore } from "pinia";
import { Session } from "@admin/api/client";
import * as apiSession from "../api/sessions";

export const useSessionsStore = defineStore("sessions", {
  state: () => ({
    sessions: [] as Array<Session>,
    session: {} as Session,
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
        this.sessions = res.data;
        this.numberSessions = parseInt(res.headers["x-total-count"], 10);
        return res;
      }

      return false;
    },

    async get(uid: string) {
      const res = await apiSession.getSession(uid);
      this.session = res.data;
    },

    clearListSessions() {
      this.sessions = [];
      this.numberSessions = 0;
    },

    clearObjectSession() {
      this.session = {} as Session;
    },
  },
});

export default useSessionsStore;
