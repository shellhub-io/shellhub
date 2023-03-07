import { Module } from "vuex";

export interface BoxState {
  status: boolean;
}

export function createBoxModule() {
  const box: Module<BoxState, any> = {
    namespaced: true,
    state: {
      status: true,
    },

    getters: {
      getStatus: (state) => state.status,
    },

    mutations: {
      setStatus: (state, status) => {
        state.status = status;
      },
    },

    actions: {
      setStatus: async (context, status) => {
        context.commit("setStatus", status);
      },
    },
  };

  return box;
}
