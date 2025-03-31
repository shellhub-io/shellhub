import { Module } from "vuex";
import * as apiUser from "../api/users";
import { State } from "..";

export interface SecurityState {
  sessionRecord: boolean,

}

export const security: Module<SecurityState, State> = {
  namespaced: true,
  state: {
    sessionRecord: true,
  },

  getters: {
    get: (state) => state.sessionRecord,
  },

  mutations: {
    setSecurity: (state, res) => {
      state.sessionRecord = res;
    },
  },

  actions: {
    async set(context, data) {
      await apiUser.setSessionRecordStatus(data);
      context.commit("setSecurity", data.status);
    },

    async get(context) {
      const res = await apiUser.getSessionRecordStatus();
      context.commit("setSecurity", res.data);
    },
  },
};
