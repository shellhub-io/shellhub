import Vue from 'vue';
import * as apiDevice from '@/api/devices';

export default {
  namespaced: true,

  state: {
    devices: [],
    device: [],
    numberDevices: 0,
    page: 0,
    perPage: 0,
    filter: null,
    status: '',
  },

  getters: {
    list: (state) => state.devices,
    get: (state) => state.device,
    getNumberDevices: (state) => state.numberDevices,
    getStatusCode: (state) => state.statusCode,
    getPage: (state) => state.page,
    getPerPage: (state) => state.perPage,
    getFilter: (state) => state.filter,
    getStatus: (state) => state.status,
  },

  mutations: {
    setDevices: (state, res) => {
      Vue.set(state, 'devices', res.data);
      Vue.set(state, 'numberDevices', parseInt(res.headers['x-total-count'], 10));
    },

    removeDevice: (state, uid) => {
      state.devices.splice(state.devices.findIndex((d) => d.uid === uid), 1);
    },

    renameDevice: (state, data) => {
      Vue.set(state, 'devices', state.devices.map((i) => (i.uid === data.uid ? { ...i, name: data.name } : i)));
    },

    setDevice: (state, data) => {
      if (data) {
        Vue.set(state, 'device', data);
      }
    },

    setPagePerpageFilter: (state, data) => {
      Vue.set(state, 'page', data.page);
      Vue.set(state, 'perPage', data.perPage);
      Vue.set(state, 'filter', data.filter);
      Vue.set(state, 'status', data.status);
    },

    setFilter: (state, filter) => {
      Vue.set(state, 'filter', filter);
    },
  },

  actions: {
    fetch: async (context, data) => {
      const res = await apiDevice.fetchDevices(data.perPage, data.page, data.filter, data.status);
      context.commit('setDevices', res);
      context.commit('setPagePerpageFilter', data);
    },

    remove: async (context, uid) => {
      await apiDevice.removeDevice(uid);
    },

    rename: async (context, data) => {
      await apiDevice.renameDevice(data);
      context.commit('renameDevice', data);
    },

    get: async (context, uid) => {
      const res = await apiDevice.getDevice(uid);
      context.commit('setDevice', res.data);
    },

    accept: async (context, uid) => {
      await apiDevice.acceptDevice(uid);
    },

    reject: async (context, uid) => {
      await apiDevice.rejectDevice(uid);
    },

    setFilter: async (context, filter) => {
      context.commit('setFilter', filter);
    },

    refresh: async ({ commit, state }) => {
      const res = await apiDevice.fetchDevices(
        state.perPage,
        state.page,
        state.filter,
        state.status,
      );
      commit('setDevices', res);
    },
  },
};
