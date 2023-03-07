import { Module } from "vuex";

export interface MobileState {
  isMobile: boolean;
}

export function createMobileModule() {
  const mobile: Module<MobileState, any> = {
    namespaced: true,
    state: {
      isMobile: false,
    },

    getters: {
      isMobile: (state) => state.isMobile,
    },

    mutations: {
      setIsMobileStatus: (state, status) => {
        state.isMobile = status;
      },
    },

    actions: {
      setIsMobileStatus(context, status) {
        context.commit("setIsMobileStatus", status);
      },
    },
  };

  return mobile;
}
