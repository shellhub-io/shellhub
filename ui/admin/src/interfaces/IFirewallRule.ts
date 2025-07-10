import { FirewallRuleFilter, IFirewallRule } from "@/interfaces/IFirewallRule";

export type AdminFirewallRuleFilter = FirewallRuleFilter;

export type IAdminFirewallRule = Omit<IFirewallRule, "status">;
