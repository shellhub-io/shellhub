import { adminApi } from "./../../api/http";

const fetchFirewalls = async (page : number, perPage: number) => adminApi.getFirewallRulesAdmin(page, perPage);

const getFirewall = async (id: number) => adminApi.getFirewallRuleAdmin(id);

export { fetchFirewalls, getFirewall };
