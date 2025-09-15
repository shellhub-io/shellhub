import { adminApi } from "@admin/api/http";

export const fetchFirewalls = async (page: number, perPage: number) => adminApi.getFirewallRulesAdmin(page, perPage);

export const getFirewall = async (id: string) => adminApi.getFirewallRuleAdmin(id);
