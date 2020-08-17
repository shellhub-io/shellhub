import Vue from 'vue';
import * as apiDevice from '@/store/api/devices';

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

    clearListNotifications: (state) => {
      Vue.set(state, 'notifications', []);
      Vue.set(state, 'numberNotifications', 0);
    },
  },

  actions: {
    fetch: async (context) => {
      try {
        const res = await apiDevice.fetchDevices(10, 1, null, 'pending', null, '');
        context.commit('setNotifications', res);
      } catch (error) {
        context.commit('clearListNotifications');
        throw error;
      }
    },
  },
};
