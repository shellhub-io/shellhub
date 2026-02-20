import { defineStore } from "pinia";
import { ref } from "vue";
import * as firewallRuleApi from "../api/firewall_rules";
import { IFirewallRule } from "@/interfaces/IFirewallRule";
import { parseTotalCount } from "@/utils/headers";

const useFirewallRulesStore = defineStore("firewallRules", () => {
  const firewallRules = ref<Array<IFirewallRule>>([]);
  const firewallRuleCount = ref<number>(0);

  const createFirewallRule = async (data: IFirewallRule) => {
    await firewallRuleApi.createFirewallRule(data);
  };

  const fetchFirewallRuleList = async (data?: { perPage: number; page: number }) => {
    try {
      const res = await firewallRuleApi.fetchFirewallRuleList(
        data?.perPage || 10,
        data?.page || 1,
      );
      firewallRules.value = res.data as IFirewallRule[];
      firewallRuleCount.value = parseTotalCount(res.headers);
    } catch (error) {
      firewallRules.value = [];
      firewallRuleCount.value = 0;
      throw error;
    }
  };

  const updateFirewallRule = async (data: IFirewallRule) => {
    await firewallRuleApi.updateFirewallRule(data);
  };

  const removeFirewallRule = async (id: string) => {
    await firewallRuleApi.removeFirewallRule(id);
  };

  return {
    firewallRules,
    firewallRuleCount,

    createFirewallRule,
    fetchFirewallRuleList,
    updateFirewallRule,
    removeFirewallRule,
  };
});

export default useFirewallRulesStore;
