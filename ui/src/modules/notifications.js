import Vue from 'vue';
import * as apiDevice from '@/api/devices';

export default {
  namespaced: true,

  state: {
    notifications: [],
    numberNotifications: 0,
  },

  getters: {
    list: (state) => state.notifications,
    getNumberNotifications: (state) => state.numberNotifications,
  },

  mutations: {
    setNotifications: (state, res) => {
      Vue.set(state, 'notifications', res.data);
      Vue.set(state, 'numberNotifications', parseInt(res.headers['x-total-count'], 10));
    },
  },

  actions: {
    fetch: async (context) => {
      const res = await apiDevice.fetchDevices(10, 1, null, 'pending');
      context.commit('setNotifications', res);
    },
  },
};
