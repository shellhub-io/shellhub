import { Module } from "vuex";
import { useChatWoot } from "@productdevbook/chatwoot/vue";
import * as apiSupport from "../api/namespaces";
import { State } from "..";

const { toggle } = useChatWoot();

export interface SupportState {
  identifier: string;
}

export const support: Module<SupportState, State> = {
  namespaced: true,
  state: {
    identifier: "",
  },

  getters: {
    getIdentifier: (state) => state.identifier,
  },

  mutations: {
    setIdentifier: (state, identifier) => {
      state.identifier = identifier;
    },
  },

  actions: {
    get: async (context, data) => {
      try {
        const res = await apiSupport.getSupportID(data);
        context.commit("setIdentifier", res.data.identifier);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },
    toggle: () => { toggle("open"); },
  },
};
