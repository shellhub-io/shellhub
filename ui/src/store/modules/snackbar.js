import Vue from 'vue';

export default {
  namespaced: true,

  state: {
    snackbarError: false,
    snackbarSuccess: false,
    snackbarMessageAndContentType: { typeMessage: '', typeContent: '' },
    snackbarCopy: false,
  },

  getters: {
    snackbarSuccess: (state) => state.snackbarSuccess,
    snackbarError: (state) => state.snackbarError,
    snackbarMessageAndContentType: (state) => state.snackbarMessageAndContentType,
    snackbarCopy: (state) => state.snackbarCopy,
  },

  mutations: {
    setSnackbarSuccessAction: (state, data) => {
      Vue.set(state, 'snackbarMessageAndContentType', data);
      Vue.set(state, 'snackbarSuccess', true);
    },

    setSnackbarSuccessDefault: (state) => {
      Vue.set(state, 'snackbarMessageAndContentType', { typeMessage: 'default', typeContent: '' });
      Vue.set(state, 'snackbarSuccess', true);
    },

    unsetSnackbarSuccess: (state) => {
      Vue.set(state, 'snackbarSuccess', false);
    },

    setSnackbarErrorLoadingOrAction: (state, data) => {
      Vue.set(state, 'snackbarMessageAndContentType', data);
      Vue.set(state, 'snackbarError', true);
    },

    setSnackbarErrorDefault: (state) => {
      Vue.set(state, 'snackbarMessageAndContentType', { typeMessage: 'default', typeContent: '' });
      Vue.set(state, 'snackbarError', true);
    },

    setSnackbarErrorAssociation: (state, data) => {
      Vue.set(state, 'snackbarMessageAndContentType', data);
      Vue.set(state, 'snackbarError', true);
    },

    unsetSnackbarError: (state) => {
      Vue.set(state, 'snackbarError', false);
    },

    setSnackbarCopy: (state, value) => {
      Vue.set(state, 'snackbarMessageAndContentType', { typeMessage: '', typeContent: value });
      Vue.set(state, 'snackbarCopy', true);
    },

    unsetSnackbarCopy: (state) => {
      Vue.set(state, 'snackbarCopy', false);
    },

    setSnackbarSuccessNotRequest: (state, data) => {
      Vue.set(state, 'snackbarMessageAndContentType', data);
      Vue.set(state, 'snackbarSuccess', true);
    },

    setSnackbarErrorNotRequest: (state, data) => {
      Vue.set(state, 'snackbarMessageAndContentType', data);
      Vue.set(state, 'snackbarError', true);
    },
  },

  actions: {
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

    showSnackbarErrorAssociation: (context) => {
      const data = { typeMessage: 'association', typeContent: '' };
      context.commit('setSnackbarErrorAssociation', data);
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

    showSnackbarSuccessNotRequest: (context, value) => {
      const data = { typeMessage: 'notRequest', typeContent: value };
      context.commit('setSnackbarSuccessNotRequest', data);
    },

    showSnackbarErrorNotRequest: (context, value) => {
      const data = { typeMessage: 'notRequest', typeContent: value };
      context.commit('setSnackbarErrorNotRequest', data);
    },
  },
};
