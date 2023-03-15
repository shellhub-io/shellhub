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
      try {
        await apiUser.putSecurity(data);
        context.commit("setSecurity", data.status);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    async get(context) {
      try {
        const res = await apiUser.getSecurity();
        context.commit("setSecurity", res.data);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },
  },
};
