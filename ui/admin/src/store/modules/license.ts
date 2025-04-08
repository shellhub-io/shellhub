import { Module } from "vuex";
import { State } from "./../index";
import { ILicense } from "./../../interfaces/ILicense";
import * as apiLicense from "../api/license";

export interface LicenseState {
  license: ILicense;
}

export const license: Module<LicenseState, State> = {
  namespaced: true,

  getters: {
    isExpired: (state) => (state.license && state.license.expired)
      || (state.license && state.license.expired === undefined),
    license: (state) => state.license,
  },

  mutations: {
    setLicense: (state, res) => {
      state.license = res.data;
    },
  },

  actions: {
    async get({ commit }) {
      const res = await apiLicense.getLicense();
      commit("setLicense", res);
    },

    async post(context, file: File) {
      await apiLicense.uploadLicense(file);
    },
  },
};
