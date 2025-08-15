import { defineStore } from "pinia";
import { ref } from "vue";
import * as firewallRuleApi from "../api/firewall_rules";
import { IFirewallRule } from "@/interfaces/IFirewallRule";

const useFirewallRulesStore = defineStore("firewallRules", () => {
  const firewallRules = ref<Array<IFirewallRule>>([]);
  const firewall = ref<IFirewallRule>({} as IFirewallRule);
  const firewallRuleCount = ref<number>(0);

  const createFirewallRule = async (data: IFirewallRule) => {
    await firewallRuleApi.postFirewall(data);
  };

  const fetchFirewallRuleList = async (data?: { perPage: number; page: number }) => {
    try {
      const res = await firewallRuleApi.fetchFirewalls(
        data?.perPage || 10,
        data?.page || 1,
      );
      if (res.data.length) {
        firewallRules.value = res.data as IFirewallRule[];
        firewallRuleCount.value = parseInt(res.headers["x-total-count"], 10);
        return true;
      }
      return false;
    } catch (error) {
      firewallRules.value = [];
      firewallRuleCount.value = 0;
      throw error;
    }
  };

  // const refresh = async () => {
  //   try {
  //     const res = await firewallRuleApi.fetchFirewalls(
  //       perPage.value,
  //       page.value,
  //     );
  //     firewalls.value = res.data;
  //     numberFirewalls.value = parseInt(res.headers["x-total-count"], 10);
  //   } catch (error) {
  //     firewalls.value = [];
  //     numberFirewalls.value = 0;
  //     throw error;
  //   }
  // };

  const updateFirewallRule = async (data: IFirewallRule) => {
    await firewallRuleApi.updateFirewallRule(data);
  };

  const removeFirewallRule = async (id: string) => {
    await firewallRuleApi.removeFirewallRule(id);
  };

  const setFirewall = (firewallData: IFirewallRule) => {
    firewall.value = firewallData;
  };

  const removeFromList = (id: string | number) => {
    const index = firewallRules.value.findIndex((d) => d.id === id);
    if (index !== -1) {
      firewallRules.value.splice(index, 1);
      firewallRuleCount.value = Math.max(0, firewallRuleCount.value - 1);
    }
  };

  const clearFirewalls = () => {
    firewallRules.value = [];
    firewallRuleCount.value = 0;
  };

  const clearFirewall = () => {
    firewall.value = {} as IFirewallRule;
  };

  return {
    firewallRules,
    firewall,
    firewallRuleCount,

    createFirewallRule,
    fetchFirewallRuleList,
    updateFirewallRule,
    removeFirewallRule,
    setFirewall,
    removeFromList,
    clearFirewalls,
    clearFirewall,
  };
});

export default useFirewallRulesStore;
