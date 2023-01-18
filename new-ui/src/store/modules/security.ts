import { Module } from "vuex";
import { State } from "./../index";
import * as apiUser from '../api/users';


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
        context.commit('setSecurity', data.status);
      } catch (error) {
        throw error;
      }
    },

    async get(context) {
      try {
        const res = await apiUser.getSecurity();
        context.commit('setSecurity', res.data);
      } catch (error) {
        throw error;
      }
    },
  },
};
