import { Module } from "vuex";
import { useChatWoot } from "@productdevbook/chatwoot/vue";
import * as apiSupport from "../api/namespaces";
import { State } from "..";

export interface SupportState {
  identifier: string;
  chatCreated: boolean;
}

export const support: Module<SupportState, State> = {
  namespaced: true,
  state: {
    identifier: "",
    chatCreated: false,
  },

  getters: {
    getIdentifier: (state) => state.identifier,
    getCreatedStatus: (state) => state.chatCreated,
  },

  mutations: {
    setIdentifier: (state, identifier) => {
      state.identifier = identifier;
    },

    setCreatedStatus: (state, status) => {
      state.chatCreated = status;
    },
  },

  actions: {
    get: async (context, data) => {
      useChatWoot().reset();
      const res = await apiSupport.getSupportID(data);
      context.commit("setIdentifier", res.data.identifier);
    },
  },
};
