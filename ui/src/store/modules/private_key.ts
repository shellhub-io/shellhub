/* eslint-disable */
import { Module } from "vuex";
import { State } from "./../index";
import { IPrivateKey } from "@/interfaces/IPrivateKey";

export interface PrivateKeyState {
  privateKeys: Array<IPrivateKey>,
  numberPrivateKeys: number,
}

export const privateKey: Module<PrivateKeyState, State> = {
  namespaced: true,
  state: {
    privateKeys: [],
    numberPrivateKeys: 0,
  },

  getters: {
    list: (state) => state.privateKeys,
    getNumberPrivateKeys: (state) => state.numberPrivateKeys,
  },

  mutations: {
    fetchPrivateKey: (state, privateKey) => {
      state.privateKeys = privateKey;
      state.numberPrivateKeys = privateKey.length;
    },

    setPrivateKey: (state, privateKey) => {
      let { numberPrivateKeys } = state;

      state.privateKeys.push(privateKey);
      state.numberPrivateKeys = numberPrivateKeys += 1;
    },

    editPrivateKey: (state, data) => {
      const { index, ...pk } = data;
      state.privateKeys.splice(index, 1, pk);
      localStorage.setItem("privateKeys", JSON.stringify(state.privateKeys));
    },

    removePrivateKey: (state, data) => {
      state.privateKeys.splice(state.privateKeys.findIndex((d) => d.data === data), 1);
      state.numberPrivateKeys = state.privateKeys.length;
    },
  },

  actions: {
    fetch: async (context) => {
      // @ts-ignore
      const privateKeys = JSON.parse(localStorage.getItem("privateKeys"));
      if (privateKeys !== null) { context.commit("fetchPrivateKey", privateKeys); }
    },

    set: async (context, privateKey) => {
      // @ts-ignore
      const privateKeys = JSON.parse(localStorage.getItem("privateKeys")) || [];

      privateKeys.forEach((pk : any) => {
        if (pk.data === privateKey.data && pk.name === privateKey.name) {
          throw new Error("both");
        }
        if (pk.data === privateKey.data) {
          throw new Error("private_key");
        }
        if (pk.name === privateKey.name) {
          throw new Error("name");
        }
      });
      privateKeys.push(privateKey);
      localStorage.setItem("privateKeys", JSON.stringify(privateKeys));
      context.commit("setPrivateKey", privateKey);
    },

    edit: async (context, privateKey) => {
      let index;
      context.state.privateKeys.forEach((pk, i) => {
        if (pk.data === privateKey.data) {
          index = i;
        }
        if (pk.name === privateKey.name) {
          throw new Error("name");
        }
      });
      context.commit("editPrivateKey", { ...privateKey, ...{ index } });
    },

    remove: async (context, data) => {
      // @ts-ignore
      const privateKeys = JSON.parse(localStorage.getItem("privateKeys")) || [];

      if (privateKeys !== null) {
        privateKeys.splice(privateKeys.findIndex((d: any) => d.data === data), 1);
      }

      localStorage.setItem("privateKeys", JSON.stringify(privateKeys));
      context.commit("removePrivateKey", data);
    },
  },
};
