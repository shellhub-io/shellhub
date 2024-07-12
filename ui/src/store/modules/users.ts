import { Module } from "vuex";
import * as apiUser from "../api/users";
import { State } from "..";

export interface UsersState {
  statusUpdateAccountDialog: boolean;
  statusUpdateAccountDialogByDeviceAction: boolean;
  deviceDuplicationError: boolean,
  showPaywall: boolean,
  premiumContent: Array<object>,
}

export const users: Module<UsersState, State> = {
  namespaced: true,
  state: {
    statusUpdateAccountDialog: false,
    statusUpdateAccountDialogByDeviceAction: false,
    deviceDuplicationError: false,
    showPaywall: false,
    premiumContent: [],
  },

  getters: {
    statusUpdateAccountDialog: (state) => state.statusUpdateAccountDialog,
    statusUpdateAccountDialogByDeviceAction(state) {
      return state.statusUpdateAccountDialogByDeviceAction;
    },
    deviceDuplicationError: (state) => state.deviceDuplicationError,
    showPaywall: (state) => state.showPaywall,
    getPremiumContent: (state) => state.premiumContent,
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

    setShowPaywall(state, status) {
      state.showPaywall = status;
    },
    setPremiumContent(state, data) {
      state.premiumContent = data;
    },
  },

  actions: {
    async signUp(context, data) {
      try {
        await apiUser.signUp(data);
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

    async getPremiumContent(context) {
      try {
        const res = await apiUser.premiumContent();
        context.commit("setPremiumContent", res);
      } catch (error) {
        console.error(error);
        throw error;
      }
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
