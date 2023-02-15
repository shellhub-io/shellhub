import { Module } from "vuex";
import { State } from "./../index";
import * as apiFirewallRule from "../api/firewall_rules";
import { IFirewall } from "@/interfaces/IFirewall";

export interface FirewallRulesState {
  firewalls: Array<IFirewall>;
  firewall: IFirewall;
  numberFirewalls: number;
  page: number;
  perPage: number;
  filter: string | null;
}

export const firewallRules: Module<FirewallRulesState, State> = {
  namespaced: true,
  state: {
    firewalls: [],
    firewall: {} as IFirewall,
    numberFirewalls: 0,
    page: 1,
    perPage: 10,
    filter: null,
  },

  getters: {
    list: (state) => state.firewalls,
    get: (state) => state.firewall,
    getNumberFirewalls: (state) => state.numberFirewalls,
    getPage: (state) => state.page,
    getPerPage: (state) => state.perPage,
  },

  mutations: {
    setFirewalls: (state, res) => {
      state.firewalls = res.data;
      state.numberFirewalls = parseInt(res.headers["x-total-count"], 10);
    },

    setFirewall: (state, res) => {
      state.firewall = res.data;
    },

    removeFirewalls: (state, id) => {
      state.firewalls.splice(
        state.firewalls.findIndex((d) => d.id === id),
        1,
      );
    },

    setPagePerpageFilter: (state, data) => {
      state.page = data.page;
      state.perPage = data.perPage;
      state.filter = data.filter;
    },

    resetPagePerpage: (state) => {
      state.page = 1;
      state.perPage = 10;
    },

    clearListFirewalls: (state) => {
      state.firewalls = [];
      state.numberFirewalls = 0;
    },

    clearObjectFirewalls: (state) => {
      state.firewall = {} as IFirewall;
    },
  },

  actions: {
    post: async (context, data) => {
      try {
        await apiFirewallRule.postFirewall(data);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    fetch: async (context, data) => {
      try {
        const res = await apiFirewallRule.fetchFirewalls(
          data.perPage,
          data.page,
        );
        if (res.data.length) {
          context.commit("setFirewalls", res);
          context.commit("setPagePerpageFilter", data);
          return true;
        }
        return false;
      } catch (error) {
        context.commit("clearListFirewalls");
        throw error;
      }
    },

    refresh: async (context) => {
      try {
        const res = await apiFirewallRule.fetchFirewalls(
          context.state.perPage,
          context.state.page,
        );
        context.commit("setFirewalls", res);
      } catch (error) {
        context.commit("clearListFirewalls");
        throw error;
      }
    },

    get: async (context, id) => {
      try {
        const res = await apiFirewallRule.getFirewall(id);
        context.commit("setFirewall", res);
      } catch (error) {
        context.commit("clearObjectFirewalls");
        throw error;
      }
    },

    put: async (context, data) => {
      await apiFirewallRule.putFirewall(data);
    },

    resetPagePerpage: async (context) => {
      context.commit("resetPagePerpage");
    },

    remove: async (context, id) => {
      try {
        await apiFirewallRule.removeFirewall(id);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },
  },
};
