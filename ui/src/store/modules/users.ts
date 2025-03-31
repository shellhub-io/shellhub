import { Module } from "vuex";
import { AxiosResponse } from "axios";
import * as apiUser from "../api/users";
import { State } from "..";

export interface UsersState {
  statusUpdateAccountDialog: boolean;
  statusUpdateAccountDialogByDeviceAction: boolean;
  deviceDuplicationError: boolean,
  showPaywall: boolean,
  premiumContent: Array<object>,
  signUpToken: string | undefined,
  info: {
    version: string;
    endpoints: {
      ssh: string;
      api: string;
    };
    setup: boolean;
    authentication: {
      local: boolean,
      saml: boolean
    }
  },
  samlUrl: string,
}

export const users: Module<UsersState, State> = {
  namespaced: true,
  state: {
    statusUpdateAccountDialog: false,
    statusUpdateAccountDialogByDeviceAction: false,
    deviceDuplicationError: false,
    showPaywall: false,
    premiumContent: [],
    signUpToken: undefined,
    info: {
      version: "",
      endpoints: {
        ssh: "",
        api: "",
      },
      setup: false,
      authentication: {
        local: false,
        saml: false,
      },
    },
    samlUrl: "",
  },

  getters: {
    statusUpdateAccountDialog: (state) => state.statusUpdateAccountDialog,
    statusUpdateAccountDialogByDeviceAction(state) {
      return state.statusUpdateAccountDialogByDeviceAction;
    },
    deviceDuplicationError: (state) => state.deviceDuplicationError,
    showPaywall: (state) => state.showPaywall,
    getPremiumContent: (state) => state.premiumContent,
    getSignToken: (state) => state.signUpToken,
    getSystemInfo: (state) => state.info,
    getSamlURL: (state) => state.samlUrl,
  },

  mutations: {
    updateStatusUpdateAccountDialog(state, status) {
      state.statusUpdateAccountDialog = status;
    },

    updateStatusUpdateAccountDialogByDeviceAction(state, status) {
      state.statusUpdateAccountDialogByDeviceAction = status;
    },
    updateDeviceDuplicationError(state, status) {
      state.deviceDuplicationError = status;
    },

    setSignUpToken(state, token) {
      state.signUpToken = token;
    },

    setShowPaywall(state, status) {
      state.showPaywall = status;
    },

    setPremiumContent(state, data) {
      state.premiumContent = data;
    },

    setSystemInfo(state, data) {
      state.info = data;
    },

    setSamlUrl(state, url) {
      state.samlUrl = url;
    },
  },

  actions: {
    async signUp(context, data) {
      const res: AxiosResponse = await apiUser.signUp(data);

      if (res.data.token) {
        context.commit("setSignUpToken", res.data.token);
        context.commit("auth/authSuccess", res.data, { root: true });
        localStorage.setItem("token", res.data.token || "");
        localStorage.setItem("user", res.data.user || "");
        localStorage.setItem("name", res.data.name || "");
        localStorage.setItem("tenant", res.data.tenant || "");
        localStorage.setItem("email", res.data.email || "");
        localStorage.setItem("id", res.data.id || "");
        localStorage.setItem("role", res.data.role || "");
        localStorage.setItem("namespacesWelcome", JSON.stringify({}));
      }
    },

    async patchData(context, data) {
      await apiUser.patchUserData(data);
    },

    async setup(context, data) {
      await apiUser.setup(data);
    },

    async patchPassword(context, data) {
      await apiUser.patchUserPassword(data);
    },

    async resendEmail(context, username) {
      await apiUser.postResendEmail(username);
    },

    async recoverPassword(context, username) {
      await apiUser.postRecoverPassword(username);
    },

    async validationAccount(context, data) {
      await apiUser.postValidationAccount(data);
    },

    async updatePassword(context, data) {
      await apiUser.postUpdatePassword(data);
    },

    async getPremiumContent(context) {
      const res = await apiUser.premiumContent();
      context.commit("setPremiumContent", res);
    },

    async fetchSystemInfo({ commit }) {
      const response = await apiUser.getInfo();
      commit("setSystemInfo", response.data);
    },

    async fetchSamlUrl({ commit }) {
      const response = await apiUser.getSamlLink();
      commit("setSamlUrl", response.data.url);
    },

    setStatusUpdateAccountDialog(context, status) {
      context.commit("updateStatusUpdateAccountDialog", status);
    },

    setStatusUpdateAccountDialogByDeviceAction(context, status) {
      context.commit("updateStatusUpdateAccountDialogByDeviceAction", status);
    },

    setDeviceDuplicationOnAcceptance(context, status) {
      context.commit("updateDeviceDuplicationError", status);
    },
  },
};
