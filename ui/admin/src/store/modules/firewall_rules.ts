import { defineStore } from "pinia";
import { ref } from "vue";
import { IAdminFirewallRule } from "@admin/interfaces/IFirewallRule";
import * as firewallRulesApi from "../api/firewall_rules";

const useFirewallRulesStore = defineStore("firewallRules", () => {
  const firewallRules = ref<Array<IAdminFirewallRule>>([]);
  const firewallRulesCount = ref(0);

  const fetchFirewallRulesList = async (data?: { page: number; perPage: number }) => {
    const res = await firewallRulesApi.fetchFirewalls(data?.page || 1, data?.perPage || 10);
    firewallRules.value = res.data as IAdminFirewallRule[];
    firewallRulesCount.value = parseInt(res.headers["x-total-count"], 10);
  };

  const fetchFirewallRuleById = async (uid: string) => {
    const res = await firewallRulesApi.getFirewall(uid);
    return res.data as IAdminFirewallRule;
  };

  return {
    firewallRules,
    firewallRulesCount,
    fetchFirewallRulesList,
    fetchFirewallRuleById,
  };
});

export default useFirewallRulesStore;
