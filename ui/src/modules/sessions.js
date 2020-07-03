import Vue from 'vue';
import * as apiSession from '@/api/sessions';

export default {
  namespaced: true,

  state: {
    sessions: [],
    session: [],
    numberSessions: 0,
    logSession: [],
  },

  getters: {
    list: (state) => state.sessions,
    get: (state) => state.session,
    getNumberSessions: (state) => state.numberSessions,
    getLogSession: (state) => state.logSession,
  },

  mutations: {
    setSessions: (state, res) => {
      Vue.set(state, 'sessions', res.data);
      Vue.set(state, 'numberSessions', parseInt(res.headers['x-total-count'], 10));
    },
    setSession: (state, data) => {
      if (data) {
        Vue.set(state, 'session', data);
        Vue.set(state, 'logSession', data);
      }
    },
  },

  actions: {
    fetch: async (context, data) => {
      const res = await apiSession.fetchSessions(data.perPage, data.page);
      context.commit('setSessions', res);
    },
    get: async (context, uid) => {
      const res = await apiSession.getSession(uid);
      context.commit('setSession', res.data);
    },
    close: async (context, session) => {
      await apiSession.closeSession(session);
    },
    getLogSession: async (context, uid) => {
      const res = await apiSession.getLog(uid);
      context.commit('setSession', res.data);
    },
  },
};
