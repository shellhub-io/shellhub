import { Module } from "vuex";
import { State } from "..";

export interface snackbarMessageAndContentType {
  typeMessage: string;
  typeContent: string;
}

export interface SnackbarState {
  message: string,
  type: string,
  snackbarError: boolean;
  snackbarSuccess: boolean;
  snackbarMessageAndContentType: snackbarMessageAndContentType;
  snackbarCopy: boolean;
}

export const snackbar: Module<SnackbarState, State> = {
  namespaced: true,
  state: {
    message: "",
    type: "",
    snackbarError: false,
    snackbarSuccess: false,
    snackbarMessageAndContentType: {
      typeMessage: "",
      typeContent: "",
    },
    snackbarCopy: false,
  },

  getters: {
    snackbarSuccess: (state) => state.snackbarSuccess,
    snackbarError: (state) => state.snackbarError,
    snackbarMessageAndContentType: (state) => state.snackbarMessageAndContentType,
    snackbarCopy: (state) => state.snackbarCopy,
  },

  mutations: {
    showMessage: (state, data) => {
      state.type = data.type;
      state.message = data.message;
    },

    setSnackbarSuccessAction: (state, data) => {
      state.snackbarMessageAndContentType = data;
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
      state.snackbarMessageAndContentType = {
        typeMessage: "default",
        typeContent: "",
      };
      state.snackbarError = true;
    },

    unsetSnackbarError: (state) => {
      state.snackbarError = false;
    },

    setSnackbarCopy: (state, value) => {
      state.snackbarMessageAndContentType = {
        typeMessage: "",
        typeContent: value,
      };
      state.snackbarCopy = true;
    },

    unsetSnackbarCopy: (state) => {
      state.snackbarCopy = false;
    },
  },

  actions: {
    showSnackbarSuccessAction: ({ commit }, value) => {
      const data = {
        typeMessage: "action",
        typeContent: value,
      };
      commit("setSnackbarSuccessAction", data);
    },

    unsetShowStatusSnackbarSuccess: ({ commit }) => {
      commit("unsetSnackbarSuccess");
    },

    showSnackbarSuccessNotRequest: (context, value) => {
      const data = {
        typeMessage: "notRequest",
        typeContent: value,
      };
      context.commit("setSnackbarSuccessNotRequest", data);
    },

    showSnackbarErrorLoading: ({ commit }, value) => {
      const data = {
        typeMessage: "loading",
        typeContent: value,
      };
      commit("setSnackbarErrorLoadingOrAction", data);
    },

    showSnackbarErrorAction: ({ commit }, value) => {
      const data = {
        typeMessage: "action",
        typeContent: value,
      };
      commit("setSnackbarErrorLoadingOrAction", data);
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
