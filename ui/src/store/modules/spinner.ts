import { Module } from "vuex";

export interface SpinnerState {
  status: boolean;
}

export function createSpinnerModule() {
  const spinner: Module<SpinnerState, any> = {
    namespaced: true,
    state: {
      status: false,
    },
    getters: {
      status: (state: SpinnerState) => state.status,
    },
    mutations: {
      setStatus: (state, status) => {
        state.status = status;
      },
    },

    actions: {
      setStatus: async ({ commit }, status) => {
        commit("setStatus", status);
      },
    },
  };
  return spinner;
}
