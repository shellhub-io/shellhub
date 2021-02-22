import Vue from 'vue';
import * as apiSession from '@/store/api/sessions';

export default {
  namespaced: true,

  state: {
    sessions: [],
    session: {},
    numberSessions: 0,
    page: 0,
    perPage: 10,
  },

  getters: {
    list: (state) => state.sessions,
    get: (state) => state.session,
    getNumberSessions: (state) => state.numberSessions,
    getPage: (state) => state.page,
    getPerPage: (state) => state.perPage,
  },

  mutations: {
    setSessions: (state, res) => {
      Vue.set(state, 'sessions', res.data);
      Vue.set(state, 'numberSessions', parseInt(res.headers['x-total-count'], 10));
    },

    setSession: (state, res) => {
      Vue.set(state, 'session', res.data);
    },

    setPagePerpage: (state, data) => {
      Vue.set(state, 'page', data.page);
      Vue.set(state, 'perPage', data.perPage);
    },

    resetPagePerpage: (state) => {
      Vue.set(state, 'page', 0);
      Vue.set(state, 'perPage', 10);
    },

    clearListSessions: (state) => {
      Vue.set(state, 'sessions', []);
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
        context.commit('setPagePerpage', data);
        context.commit('setSessions', res);
      } catch (error) {
        context.commit('clearListSessions');
        throw error;
      }
    },

    refresh: async (context) => {
      try {
        const res = await apiSession.fetchSessions(
          context.state.perPage,
          context.state.page,
        );
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

    resetPagePerpage: async (context) => {
      context.commit('resetPagePerpage');
    },

    close: async (context, session) => {
      await apiSession.closeSession(session);
    },
  },
};
