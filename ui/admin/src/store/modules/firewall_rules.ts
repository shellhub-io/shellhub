// stores/firewallRules.ts
import { defineStore } from "pinia";
import { IAdminFirewallRule } from "../../interfaces/IFirewallRule";
import * as apiFirewall from "../api/firewall_rules";

export interface FirewallRulesState {
  firewalls: Array<IAdminFirewallRule>;
  firewall: IAdminFirewallRule;
  numberFirewalls: number;
}

export const useFirewallRulesStore = defineStore("firewallRules", {
  state: (): FirewallRulesState => ({
    firewalls: [],
    firewall: {} as IAdminFirewallRule,
    numberFirewalls: 0,
  }),

  getters: {
    list: (state): Array<IAdminFirewallRule> => state.firewalls,
    getFirewall: (state): IAdminFirewallRule => state.firewall,
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
