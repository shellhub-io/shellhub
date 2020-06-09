import Vue from 'vue';
import {
  postFirewall, fetchFirewalls, getFirewall, putFirewall, removeFirewall,
} from '@/api/firewall_rules';

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
    getNumberSessions: (state) => state.numberFirewalls,
  },

  mutations: {
    setFirewalls: (state, res) => {
      Vue.set(state, 'firewalls', res.data);
    },

    setFirewall: (state, res) => {
      Vue.set(state, 'firewall', res.data);
    },

    removeFirewalls: (state, id) => {
      state.firewalls.splice(state.firewalls.findIndex((d) => d.id === id), 1);
    },
  },

  actions: {
    post: async (context, data) => {
      await postFirewall(data);
    },

    fetch: async (context) => {
      const res = await fetchFirewalls();
      context.commit('setFirewalls', res);
    },

    get: async (context, id) => {
      const res = await getFirewall(id);
      context.commit('setFirewall', res);
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
