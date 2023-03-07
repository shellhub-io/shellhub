import { Module } from "vuex";

export interface ModalState {
  terminal: string;
  addDevice: boolean;
}

export function createModalModule() {
  const modal: Module<ModalState, any> = {
    namespaced: true,
    state: {
      terminal: "",
      addDevice: false,
    },

    getters: {
      terminal: (state) => state.terminal,
      addDevice: (state) => state.addDevice,
    },

    mutations: {
      setTerminal: (state, data) => {
        state.terminal = data;
      },

      setAddDevice: (state, data) => {
        state.addDevice = data;
      },
    },

    actions: {
      toggleTerminal: (context, value) => {
        context.commit("setTerminal", value);
      },

      showAddDevice: (context, value) => {
        context.commit("setAddDevice", value);
      },
    },
  };

  return modal;
}
