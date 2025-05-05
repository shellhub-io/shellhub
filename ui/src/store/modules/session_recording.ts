import { Module } from "vuex";
import * as usersApi from "../api/users";
import { State } from "..";

export interface SessionRecordingState {
  enabled: boolean,
}

export const sessionRecording: Module<SessionRecordingState, State> = {
  namespaced: true,
  state: {
    enabled: true,
  },

  getters: {
    isEnabled: (state) => state.enabled,
  },

  mutations: {
    setEnabled: (state, status) => {
      state.enabled = status;
    },
  },

  actions: {
    async setStatus(context, data) {
      await usersApi.setSessionRecordStatus(data);
      context.commit("setEnabled", data.status);
    },

    async getStatus(context) {
      const res = await usersApi.getSessionRecordStatus();
      context.commit("setEnabled", res.data);
    },
  },
};
