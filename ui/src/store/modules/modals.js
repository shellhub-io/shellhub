import Vue from 'vue';

export default {
  namespaced: true,

  state: {
    terminal: '',
    addDevice: false,
    snackbarError: false,
    snackbarSuccess: false,
  },

  getters: {
    terminal: (state) => state.terminal,
    addDevice: (state) => state.addDevice,
    snackbarSuccess: (state) => state.snackbarSuccess,
    snackbarError: (state) => state.snackbarError,
  },

  mutations: {
    setTerminal: (state, data) => {
      Vue.set(state, 'terminal', data);
    },

    setAddDevice: (state, data) => {
      Vue.set(state, 'addDevice', data);
    },

    setSnackbarSuccess: (state, data) => {
      Vue.set(state, 'snackbarSuccess', data);
    },

    setSnackbarError: (state, data) => {
      Vue.set(state, 'snackbarError', data);
    },
  },

  actions: {
    toggleTerminal: (context, value) => {
      context.commit('setTerminal', value);
    },

    showAddDevice: (context, value) => {
      context.commit('setAddDevice', value);
    },

    showSnackbarError: (context, value) => {
      context.commit('setSnackbarError', value);
    },

    showSnackbarSuccess: (context, value) => {
      context.commit('setSnackbarSuccess', value);
    },
  },
};
