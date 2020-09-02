import Vue from 'vue';

export default {
  namespaced: true,

  state: {
    terminal: '',
    addDevice: false,
    snackbarError: false,
    snackbarSuccess: false,
    SnackbarMessageAndContentType: { typeMessage: '', typeContent: '' },
    snackbarCopy: false,
  },

  getters: {
    terminal: (state) => state.terminal,
    addDevice: (state) => state.addDevice,
    snackbarSuccess: (state) => state.snackbarSuccess,
    snackbarError: (state) => state.snackbarError,
    SnackbarMessageAndContentType: (state) => state.SnackbarMessageAndContentType,
    snackbarCopy: (state) => state.snackbarCopy,
  },

  mutations: {
    setTerminal: (state, data) => {
      Vue.set(state, 'terminal', data);
    },

    setAddDevice: (state, data) => {
      Vue.set(state, 'addDevice', data);
    },

    setSnackbarSuccessAction: (state, data) => {
      Vue.set(state, 'SnackbarMessageAndContentType', data);
      Vue.set(state, 'snackbarSuccess', true);
    },

    setSnackbarSuccessDefault: (state) => {
      Vue.set(state, 'snackbarSuccess', { typeMessage: 'default', typeContent: '' });
      Vue.set(state, 'snackbarSuccess', true);
    },

    unsetSnackbarSuccess: (state) => {
      Vue.set(state, 'snackbarSuccess', false);
    },

    setSnackbarErrorLoadingOrAction: (state, data) => {
      Vue.set(state, 'SnackbarMessageAndContentType', data);
      Vue.set(state, 'snackbarError', true);
    },

    setSnackbarErrorDefault: (state) => {
      Vue.set(state, 'SnackbarMessageAndContentType', { typeMessage: 'default', typeContent: '' });
      Vue.set(state, 'snackbarError', true);
    },

    unsetSnackbarError: (state) => {
      Vue.set(state, 'snackbarError', false);
    },

    setSnackbarCopy: (state, value) => {
      Vue.set(state, 'SnackbarMessageAndContentType', { typeMessage: '', typeContent: value });
      Vue.set(state, 'snackbarCopy', true);
    },

    unsetSnackbarCopy: (state) => {
      Vue.set(state, 'snackbarCopy', false);
    },
  },

  actions: {
    toggleTerminal: (context, value) => {
      context.commit('setTerminal', value);
    },

    showAddDevice: (context, value) => {
      context.commit('setAddDevice', value);
    },

    showSnackbarSuccessAction: (context, value) => {
      const data = { typeMessage: 'action', typeContent: value };
      context.commit('setSnackbarSuccessAction', data);
    },

    showSnackbarSuccessDefault: (context) => {
      context.commit('setSnackbarSuccessDefault');
    },

    unsetShowStatusSnackbarSuccess: (context) => {
      context.commit('unsetSnackbarSuccess');
    },

    showSnackbarErrorLoading: (context, value) => {
      const data = { typeMessage: 'loading', typeContent: value };
      context.commit('setSnackbarErrorLoadingOrAction', data);
    },

    showSnackbarErrorAction: (context, value) => {
      const data = { typeMessage: 'action', typeContent: value };
      context.commit('setSnackbarErrorLoadingOrAction', data);
    },

    showSnackbarErrorDefault: (context) => {
      context.commit('setSnackbarErrorDefault');
    },

    unsetShowStatusSnackbarError: (context) => {
      context.commit('unsetSnackbarError');
    },

    showSnackbarCopy: (context, value) => {
      context.commit('setSnackbarCopy', value);
    },

    unsetShowStatusSnackbarCopy: (context) => {
      context.commit('unsetSnackbarCopy');
    },
  },
};
