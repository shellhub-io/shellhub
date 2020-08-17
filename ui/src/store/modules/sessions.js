import Vue from 'vue';
import * as apiSession from '@/store/api/sessions';

export default {
  namespaced: true,

  state: {
    sessions: [],
    session: {},
    numberSessions: 0,
  },

  getters: {
    list: (state) => state.sessions,
    get: (state) => state.session,
    getNumberSessions: (state) => state.numberSessions,
  },

  mutations: {
    setSessions: (state, res) => {
      Vue.set(state, 'sessions', res.data);
      Vue.set(state, 'numberSessions', parseInt(res.headers['x-total-count'], 10));
    },

    setSession: (state, res) => {
      Vue.set(state, 'session', res.data);
    },

    clearListSessions: (state) => {
      Vue.set(state, 'devices', []);
      Vue.set(state, 'numberSessions', 0);
    },

    clearObjectSession: (state) => {
      Vue.set(state, 'session', {});
    },
  },

  actions: {
    fetch: async (context, data) => {
      try {
        const res = await apiSession.fetchSessions(data.perPage, data.page);
        context.commit('setSessions', res);
      } catch (error) {
        context.commit('clearListSessions');
        throw error;
      }
    },

    get: async (context, uid) => {
      try {
        const res = await apiSession.getSession(uid);
        context.commit('setSession', res);
      } catch (error) {
        context.commit('clearObjectSession');
        throw error;
      }
    },

    getLogSession: async (context, uid) => {
      try {
        const res = await apiSession.getLog(uid);
        context.commit('setSession', res);
      } catch (error) {
        context.commit('clearObjectSession');
        throw error;
      }
    },

    close: async (context, session) => {
      await apiSession.closeSession(session);
    },
  },
};
