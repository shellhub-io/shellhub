import { Module } from "vuex";
import { AxiosResponse } from "axios";
import * as apiUser from "../api/users";
import { State } from "..";

export interface UsersState {
  statusUpdateAccountDialog: boolean;
  statusUpdateAccountDialogByDeviceAction: boolean;
  deviceDuplicationError: boolean,
  showPaywall: boolean,
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

    setSystemInfo(state, data) {
      state.info = data;
    },

    setSamlUrl(state, url) {
      state.samlUrl = url;
    },
  },

  actions: {
    async signUp(context, data) {
      try {
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
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    async patchData(context, data) {
      try {
        await apiUser.patchUserData(data);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    async setup(context, data) {
      try {
        await apiUser.setup(data);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    async patchPassword(context, data) {
      try {
        await apiUser.patchUserPassword(data);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    async resendEmail(context, username) {
      try {
        await apiUser.postResendEmail(username);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    async recoverPassword(context, username) {
      try {
        await apiUser.postRecoverPassword(username);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    async validationAccount(context, data) {
      try {
        await apiUser.postValidationAccount(data);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    async updatePassword(context, data) {
      try {
        await apiUser.postUpdatePassword(data);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    async fetchSystemInfo({ commit }) {
      try {
        const response = await apiUser.getInfo();
        commit("setSystemInfo", response.data);
      } catch (error) {
        console.error(error);
      }
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
