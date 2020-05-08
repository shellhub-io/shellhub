import Vue from 'vue';
import { fetchSessions, getSession, closeSession } from '@/api/sessions';

export default {
  namespaced: true,

  state: {
    sessions: [],
    session: [],
    numberSessions: 0,
  },

  getters: {
    list: (state) => state.sessions,
    get: (state) => state.session,
    getNumberSessions: (state) => state.numberSessions
  },

  mutations: {
    setSessions: (state, res) => {
      Vue.set(state, 'sessions', res.data);
      Vue.set(state, 'numberSessions', parseInt(res.headers['x-total-count']));
    },
    setSession: (state, data) => {
      if(data){
        Vue.set(state, 'session', data);
      }
    }
  },

  actions: {
    fetch: async (context, data) => {
      let res = await fetchSessions(data.perPage,data.page);
      context.commit('setSessions', res);
    },
    get: async (context,uid)  => {
      let res = await getSession(uid);
      context.commit('setSession', res.data);
    },
    close: async (context, session) => {
      await closeSession(session);
    }
  }
};