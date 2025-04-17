import { Module } from "vuex";
import * as apiSession from "../api/sessions";
import { ISessions } from "@/interfaces/ISessions";
import { State } from "..";

export interface SessionsState {
  sessions: Array<ISessions>;
  session: ISessions;
  sessionLogs: string | null;
  numberSessions: number;
  page: number;
  perPage: number;
}

export const sessions: Module<SessionsState, State> = {
  namespaced: true,
  state: {
    sessions: [],
    session: {} as ISessions,
    sessionLogs: null,
    numberSessions: 0,
    page: 1,
    perPage: 10,
  },

  getters: {
    list: (state) => state.sessions,
    get: (state) => state.session,
    getLogs: (state) => state.sessionLogs,
    getNumberSessions: (state) => state.numberSessions,
    getPage: (state) => state.page,
    getPerPage: (state) => state.perPage,
  },

  mutations: {
    setSessions: (state, res) => {
      state.sessions = res.data;
      state.numberSessions = parseInt(res.headers["x-total-count"], 10);
    },

    setSession: (state, res) => {
      state.session = res.data;
    },

    setSessionLogs: (state, res) => {
      state.sessionLogs = res.data;
    },

    setPagePerpage: (state, data) => {
      state.page = data.page;
      state.perPage = data.perPage;
    },

    resetPagePerpage: (state) => {
      state.page = 1;
      state.perPage = 10;
    },

    clearListSessions: (state) => {
      state.sessions = [];
      state.numberSessions = 0;
    },

    clearSession: (state) => {
      state.session = {} as ISessions;
    },

    clearSessionLogs: (state) => {
      state.sessionLogs = null;
    },

    removeRecordedSession: (state) => {
      state.session = {
        ...state.session,
        recorded: false,
      };
    },
  },

  actions: {
    fetch: async (context, data) => {
      try {
        const res = await apiSession.fetchSessions(data.page, data.perPage);
        if (res.data.length) {
          context.commit("setPagePerpage", data);
          context.commit("setSessions", res);
          return res;
        }
        return false;
      } catch (error) {
        context.commit("clearListSessions");
        throw error;
      }
    },

    refresh: async (context) => {
      try {
        const res = await apiSession.fetchSessions(
          context.state.page,
          context.state.perPage,
        );
        context.commit("setSessions", res);
      } catch (error) {
        context.commit("clearListSessions");
        throw error;
      }
    },

    get: async (context, uid) => {
      try {
        const res = await apiSession.getSession(uid);
        context.commit("setSession", res);
      } catch (error) {
        context.commit("clearObjectSession");
        throw error;
      }
    },

    getSessionLogs: async (context, uid) => {
      try {
        const res = await apiSession.getLog(uid);
        context.commit("setSessionLogs", res);
      } catch (error) {
        context.commit("clearSessionLogs");
        throw error;
      }
    },

    resetPagePerpage: async (context) => {
      context.commit("resetPagePerpage");
    },

    close: async (context, session) => {
      await apiSession.closeSession(session);
    },

    deleteSessionLogs: async (context, uid) => {
      await apiSession.deleteSessionLogs(uid);
      context.commit("removeRecordedSession");
    },
  },
};
