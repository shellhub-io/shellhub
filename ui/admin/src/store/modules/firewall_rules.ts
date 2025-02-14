import { Module } from "vuex";
import { State } from "./../index";
import { IFirewallRule } from "./../../interfaces/IFirewallRule";
import * as apiFirewall from "../api/firewall_rules";

export interface FirewallRulesState {
  firewalls: Array<IFirewallRule>;
  firewall: IFirewallRule;
  numberFirewalls: number;
}

export const firewallRules: Module<FirewallRulesState, State> = {
  namespaced: true,

  state: {
    firewalls: [],
    firewall: {} as IFirewallRule,
    numberFirewalls: 0,
  },

  getters: {
    list: (state): Array<IFirewallRule> => state.firewalls,
    get: (state): IFirewallRule => state.firewall,
    numberFirewalls: (state): number => state.numberFirewalls,
  },

  mutations: {
    setFirewalls: (state, res) => {
      state.firewalls = res.data;
      state.numberFirewalls = parseInt(res.headers["x-total-count"], 10);
    },

    setFirewall: (state, res) => {
      state.firewall = res.data;
    },
  },

  actions: {
    async fetch({ commit }, data) {
      const res = await apiFirewall.fetchFirewalls(data.page, data.perPage);

      if (res.data.length) {
        commit("setFirewalls", res);
        return true;
      }

      return false;
    },

    async get({ commit }, uid) {
      const res = await apiFirewall.getFirewall(uid);
      commit("setFirewall", res);
    },
  },
};
