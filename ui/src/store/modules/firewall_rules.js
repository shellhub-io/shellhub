import Vue from 'vue';
import {
  postFirewall, fetchFirewalls, getFirewall, putFirewall, removeFirewall,
} from '@/store/api/firewall_rules';

export default {
  namespaced: true,

  state: {
    firewalls: [],
    firewall: {},
    numberFirewalls: 0,
    page: 0,
    perPage: 10,
    filter: null,
  },

  getters: {
    list: (state) => state.firewalls,
    get: (state) => state.firewall,
    getNumberFirewalls: (state) => state.numberFirewalls,
    getPage: (state) => state.page,
    getPerPage: (state) => state.perPage,
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

    setPagePerpageFilter: (state, data) => {
      Vue.set(state, 'page', data.page);
      Vue.set(state, 'perPage', data.perPage);
      Vue.set(state, 'filter', data.filter);
    },

    resetPagePerpage: (state) => {
      Vue.set(state, 'page', 0);
      Vue.set(state, 'perPage', 10);
    },

    clearListFirewalls: (state) => {
      Vue.set(state, 'firewalls', []);
      Vue.set(state, 'numberFirewalls', 0);
    },

    clearObjectFirewalls: (state) => {
      Vue.set(state, 'firewall', {});
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
        context.commit('setPagePerpageFilter', data);
      } catch (error) {
        context.commit('clearListFirewalls');
        throw error;
      }
    },

    refresh: async (context) => {
      try {
        const res = await fetchFirewalls(
          context.state.perPage,
          context.state.page,
        );
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

    resetPagePerpage: async (context) => {
      context.commit('resetPagePerpage');
    },

    remove: async (context, id) => {
      await removeFirewall(id);
    },
  },
};
