/* eslint-disable */
import { Module } from "vuex";
import { IPrivateKey } from "@/interfaces/IPrivateKey";
import { State } from "..";

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
      const index = state.privateKeys.findIndex((pk) => pk.id === data.id);
      state.privateKeys.splice(index, 1, data);
      localStorage.setItem("privateKeys", JSON.stringify(state.privateKeys));
    },
    

    removePrivateKey: (state, id) => {
      const index = state.privateKeys.findIndex((pk) => pk.id === id);
      if (index !== -1) {
        state.privateKeys.splice(index, 1);
        state.numberPrivateKeys = state.privateKeys.length;
        localStorage.setItem("privateKeys", JSON.stringify(state.privateKeys));
      }
    },
  },

  actions: {
      fetch: async (context) => {
        // @ts-expect-error
        let privateKeys = JSON.parse(localStorage.getItem("privateKeys")) || [];
        let maxId = 0;
        privateKeys = privateKeys.map((pk) => {
          if (!pk.id) {
            maxId += 1;
            pk.id = maxId;
          } else {
            maxId = Math.max(maxId, pk.id);
          }
          return pk;
        });
        localStorage.setItem("privateKeys", JSON.stringify(privateKeys));
        context.commit("fetchPrivateKey", privateKeys);
      },
    
    set: async (context, privateKey) => {
      // @ts-expect-error
      const privateKeys = JSON.parse(localStorage.getItem("privateKeys")) || [];
      const id = privateKeys.length ? Math.max(...privateKeys.map((pk: IPrivateKey) => pk.id)) + 1 : 1;
      privateKey.id = id;

      privateKeys.forEach((pk: IPrivateKey) => {
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
      // @ts-expect-error
      const privateKeys = JSON.parse(localStorage.getItem("privateKeys")) || [];
      const existingKey = privateKeys.find((pk: IPrivateKey) => pk.id === privateKey.id);
    
      if (existingKey && existingKey.data === privateKey.data && existingKey.name === privateKey.name) {
        throw new Error();
      }
    
      context.commit("editPrivateKey", privateKey);
    },
    

    remove: async (context, id) => {
      // @ts-expect-error
      const privateKeys = JSON.parse(localStorage.getItem("privateKeys")) || [];

      if (privateKeys !== null) {
        const index = privateKeys.findIndex((pk: IPrivateKey) => pk.id === id);
        if (index !== -1) {
          privateKeys.splice(index, 1);
        }
      }

      localStorage.setItem("privateKeys", JSON.stringify(privateKeys));
      context.commit("removePrivateKey", id);
    },
  },
};
