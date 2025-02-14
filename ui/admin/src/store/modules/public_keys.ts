import { Module } from "vuex";
import { State } from "./../index";
import {
  postPublicKey,
  fetchPublicKeys,
  getPublicKey,
  putPublicKey,
  removePublicKey,
} from "../api/public_keys";
import { IPublicKey } from "../../interfaces/IPublicKey";

export interface PublicKeysState {
  publicKeys: Array<IPublicKey>;
  publicKey: IPublicKey;
  numberPublicKeys: number;
}
export const publicKeys: Module<PublicKeysState, State> = {
  namespaced: true,
  state: {
    publicKeys: [],
    publicKey: {} as IPublicKey,
    numberPublicKeys: 0,
  },

  getters: {
    list: (state) => state.publicKeys,
    get: (state) => state.publicKey,
    getNumberPublicKeys: (state) => state.numberPublicKeys,
  },

  mutations: {
    setPublicKeys: (state, res) => {
      state.publicKeys = res.data;
      state.numberPublicKeys = parseInt(res.headers["x-total-count"], 10);
    },

    setPublicKey: (state, res) => {
      state.publicKey = res.data;
    },

    removePublicKey: (state, fingerprint) => {
      state.publicKeys.splice(
        state.publicKeys.findIndex((d: IPublicKey) => d.fingerprint === fingerprint),
        1,
      );
    },

    clearListPublicKeys: (state) => {
      state.publicKeys = [];
      state.numberPublicKeys = 0;
    },

    clearObjectPublicKey: (state) => {
      state.publicKey = {} as IPublicKey;
    },
  },

  actions: {
    post: async (context, data) => {
      await postPublicKey(data);
    },

    fetch: async ({ commit }, data) => {
      try {
        const res = await fetchPublicKeys(data.perPage, data.page);
        commit("setPublicKeys", res);
      } catch (error) {
        commit("clearListPublicKeys");
        throw error;
      }
    },

    get: async ({ commit }, id) => {
      try {
        const res = await getPublicKey(id);
        commit("setPublicKey", res);
      } catch (error) {
        commit("clearObjectPublicKey");
        throw error;
      }
    },

    put: async (context, data) => {
      await putPublicKey(data);
    },

    remove: async ({ commit }, fingerprint) => {
      await removePublicKey(fingerprint);
      commit("removePublicKey", fingerprint);
    },
  },
};
