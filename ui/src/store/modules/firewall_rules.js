import Vue from 'vue';
import {
  postFirewall, fetchFirewalls, getFirewall, putFirewall, removeFirewall,
} from '@/store/api/firewall_rules';

export default {
  namespaced: true,

  state: {
    firewalls: [],
    firewall: [],
    numberFirewalls: 0,
  },

  getters: {
    list: (state) => state.firewalls,
    get: (state) => state.firewall,
    getNumberFirewalls: (state) => state.numberFirewalls,
  },

  mutations: {
    setFirewalls: (state, res) => {
      Vue.set(state, 'firewalls', res.data);
      Vue.set(state, 'numberFirewalls', parseInt(res.headers['x-total-count'], 10));
    },

    setFirewall: (state, res) => {
      Vue.set(state, 'firewall', res.data);
    },

    removeFirewalls: (state, id) => {
      state.firewalls.splice(state.firewalls.findIndex((d) => d.id === id), 1);
    },

    clearListFirewalls: (state) => {
      Vue.set(state, 'Firewalls', []);
      Vue.set(state, 'numberFirewalls', 0);
    },

    clearObjectFirewalls: (state) => {
      Vue.set(state, 'Firewalls', []);
    },
  },

  actions: {
    post: async (context, data) => {
      await postFirewall(data);
    },

    fetch: async (context, data) => {
      try {
        const res = await fetchFirewalls(data.perPage, data.page);
        context.commit('setFirewalls', res);
      } catch (error) {
        context.commit('clearListFirewalls');
        throw error;
      }
    },

    get: async (context, id) => {
      try {
        const res = await getFirewall(id);
        context.commit('setFirewall', res);
      } catch (error) {
        context.commit('clearObjectFirewalls');
        throw error;
      }
    },

    put: async (context, data) => {
      await putFirewall(data);
    },

    remove: async (context, id) => {
      await removeFirewall(id);
      context.commit('removeFirewalls', id);
    },
  },
};
