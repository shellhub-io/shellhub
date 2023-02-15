import { Module } from "vuex";
import { State } from "./../index";
import * as apiUser from "../api/users";

export interface UsersState {
  statusUpdateAccountDialog: boolean;
  statusUpdateAccountDialogByDeviceAction: boolean;
}

export const users: Module<UsersState, State> = {
  namespaced: true,
  state: {
    statusUpdateAccountDialog: false,
    statusUpdateAccountDialogByDeviceAction: false,
  },

  getters: {
    statusUpdateAccountDialog: (state) => state.statusUpdateAccountDialog,
    statusUpdateAccountDialogByDeviceAction(state) {
      return state.statusUpdateAccountDialogByDeviceAction;
    },
  },

  mutations: {
    updateStatusUpdateAccountDialog(state, status) {
      state.statusUpdateAccountDialog = status;
    },

    updateStatusUpdateAccountDialogByDeviceAction(state, status) {
      state.statusUpdateAccountDialogByDeviceAction = status;
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

    setStatusUpdateAccountDialog(context, status) {
      context.commit("updateStatusUpdateAccountDialog", status);
    },

    setStatusUpdateAccountDialogByDeviceAction(context, status) {
      context.commit("updateStatusUpdateAccountDialogByDeviceAction", status);
    },
  },
};
