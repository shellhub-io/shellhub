import { Module } from "vuex";
import { State } from "./../index";

export interface snackbarMessageAndContentType {
  typeMessage: string;
  typeContent: string;
}

export interface SnackbarState {
  snackbarError: boolean;
  snackbarSuccess: boolean;
  snackbarMessageAndContentType: snackbarMessageAndContentType;
  snackbarCopy: boolean;
}

export const snackbar: Module<SnackbarState, State> = {
  namespaced: true,
  state: {
    snackbarError: false,
    snackbarSuccess: false,
    snackbarMessageAndContentType: { typeMessage: "", typeContent: "" },
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
      state.snackbarMessageAndContentType = data;
      state.snackbarSuccess = true;
    },

    setSnackbarNoContent: (state, data) => {
      state.snackbarMessageAndContentType = data;
      state.snackbarSuccess = true;
    },

    setSnackbarSuccessDefault: (state) => {
      state.snackbarMessageAndContentType = { typeMessage: "default", typeContent: "" };
      state.snackbarSuccess = true;
    },

    unsetSnackbarSuccess: (state) => {
      state.snackbarSuccess = false;
    },

    setSnackbarSuccessNotRequest: (state, data) => {
      state.snackbarMessageAndContentType = data;
      state.snackbarSuccess = true;
    },

    setSnackbarErrorLoadingOrAction: (state, data) => {
      state.snackbarMessageAndContentType = data;
      state.snackbarError = true;
    },

    setSnackbarErrorDefault: (state) => {
      state.snackbarMessageAndContentType = { typeMessage: "default", typeContent: "" };
      state.snackbarError = true;
    },

    setSnackbarErrorLicense: (state, data) => {
      state.snackbarMessageAndContentType = data;
      state.snackbarError = true;
    },

    unsetSnackbarError: (state) => {
      state.snackbarError = false;
    },

    setSnackbarCopy: (state, value) => {
      state.snackbarMessageAndContentType = { typeMessage: "", typeContent: value };
      state.snackbarCopy = true;
    },

    unsetSnackbarCopy: (state) => {
      state.snackbarCopy = false;
    },
  },

  actions: {
    showSnackbarSuccessAction: ({ commit }, value) => {
      const data = { typeMessage: "action", typeContent: value };
      commit("setSnackbarSuccessAction", data);
    },

    showSnackbarSuccessDefault: ({ commit }) => {
      commit("setSnackbarSuccessDefault");
    },

    unsetShowStatusSnackbarSuccess: ({ commit }) => {
      commit("unsetSnackbarSuccess");
    },

    showSnackbarSuccessNotRequest: (context, value) => {
      const data = { typeMessage: 'notRequest', typeContent: value };
      context.commit('setSnackbarSuccessNotRequest', data);
    },

    showSnackbarErrorLoading: ({ commit }, value) => {
      const data = { typeMessage: "loading", typeContent: value };
      commit("setSnackbarErrorLoadingOrAction", data);
    },

    showSnackbarNoContent: ({ commit }) => {
      const data = { typeMessage: "no-content", typeContent: "" };
      commit("setSnackbarNoContent", data);
    },

    showSnackbarErrorAction: ({ commit }, value) => {
      const data = { typeMessage: "action", typeContent: value };
      commit("setSnackbarErrorLoadingOrAction", data);
    },

    showSnackbarErrorLicense: ({ commit }, value) => {
      const data = { typeMessage: "licenseRequired", typeContent: value };
      commit("setSnackbarErrorLicense", data);
    },

    showSnackbarErrorDefault: ({ commit }) => {
      commit("setSnackbarErrorDefault");
    },

    unsetShowStatusSnackbarError: ({ commit }) => {
      commit("unsetSnackbarError");
    },

    showSnackbarCopy: ({ commit }, value) => {
      commit("setSnackbarCopy", value);
    },

    unsetShowStatusSnackbarCopy: ({ commit }) => {
      commit("unsetSnackbarCopy");
    },
  },
};
