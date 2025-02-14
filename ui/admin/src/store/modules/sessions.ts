import { Module } from "vuex";
import { State } from "./../index";
import { ISessions } from "./../../interfaces/ISession";
import * as apiSession from "../api/sessions";

export interface SessionsState {
  sessions: Array<ISessions>;
  session: ISessions;
  numberSessions: number;
}

export const sessions: Module<SessionsState, State> = {
  namespaced: true,

  state: {
    sessions: [],
    session: {} as ISessions,
    numberSessions: 0,
  },

  getters: {
    sessions: (state: SessionsState) => state.sessions,
    session: (state: SessionsState) => state.session,
    numberSessions: (state: SessionsState) => state.numberSessions,
  },

  mutations: {
    setSessions: (state, res) => {
      state.sessions = res.data;
      state.numberSessions = parseInt(res.headers["x-total-count"], 10);
    },

    setSession: (state, res) => {
      state.session = res.data;
    },

    clearListSessions: (state) => {
      state.sessions = [];
      state.numberSessions = 0;
    },

    clearObjectSession: (state) => {
      state.session = {} as ISessions;
    },
  },

  actions: {
    async fetch({ commit }, data) {
      const res = await apiSession.fetchSessions(data.perPage, data.page);
      if (res.data.length) {
        commit("setSessions", res);
        return res;
      }

      return false;
    },

    async get({ commit }, uid) {
      const res = await apiSession.getSession(uid);
      commit("setSession", res);
    },
  },
};
