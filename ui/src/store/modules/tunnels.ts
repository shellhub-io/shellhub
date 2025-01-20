import { Module } from "vuex";
import { State } from "../index";
import * as apiTunnel from "../api/tunnels";
import { ITunnel, ITunnelCreate, ITunnelDelete } from "../../interfaces/ITunnel";

export interface TunnelsState {
  tunnels: Array<ITunnel>;
}

export const tunnels: Module<TunnelsState, State> = {
  namespaced: true,
  state: {
    tunnels: [],
  },

  getters: {
    listTunnels: (state) => state.tunnels,
  },

  mutations: {
    setTunnels: (state, res) => {
      state.tunnels = res.data;
    },
  },

  actions: {
    async get({ commit }, uid) {
      const res = await apiTunnel.getTunnels(uid);
      commit("setTunnels", res);
      return res;
    },

    async delete(_, data: ITunnelDelete) {
      const { uid, address } = data;
      const res = await apiTunnel.deleteTunnel(uid, address);
      return res;
    },

    async create(_, data: ITunnelCreate) {
      const { uid, host, port, ttl } = data;
      const res = await apiTunnel.createTunnel(uid, host, port, ttl);
      return res;
    },
  },
};
