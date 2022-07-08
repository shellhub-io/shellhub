import { Module } from "vuex";
import { State } from "./../index";

export interface ModalState {
  terminal: string;
  addDevice: boolean;
}

export const modal: Module<ModalState, State> = {
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
