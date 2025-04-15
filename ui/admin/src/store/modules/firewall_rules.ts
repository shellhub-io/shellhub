// stores/firewallRules.ts
import { defineStore } from "pinia";
import { IFirewallRule } from "../../interfaces/IFirewallRule";
import * as apiFirewall from "../api/firewall_rules";

export interface FirewallRulesState {
  firewalls: Array<IFirewallRule>;
  firewall: IFirewallRule;
  numberFirewalls: number;
}

export const useFirewallRulesStore = defineStore("firewallRules", {
  state: (): FirewallRulesState => ({
    firewalls: [],
    firewall: {} as IFirewallRule,
    numberFirewalls: 0,
  }),

  getters: {
    list: (state): Array<IFirewallRule> => state.firewalls,
    getFirewall: (state): IFirewallRule => state.firewall,
    getNumberFirewalls: (state): number => state.numberFirewalls,
  },

  actions: {
    async fetch(data: { page: number; perPage: number }) {
      const res = await apiFirewall.fetchFirewalls(data.page, data.perPage);
      if (res.data.length) {
        this.firewalls = res.data as never;
        this.numberFirewalls = parseInt(res.headers["x-total-count"], 10);
        return true;
      }
      return false;
    },

    async get(uid) {
      const res = await apiFirewall.getFirewall(uid);
      this.firewall = res.data as never;
    },
  },
});

export default useFirewallRulesStore;
