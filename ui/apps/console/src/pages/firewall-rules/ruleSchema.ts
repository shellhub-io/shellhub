import { z } from "zod";
import type { FirewallRulesRequest, FirewallRulesResponse } from "@/client";

function isValidRegex(pattern: string): boolean {
  try {
    new RegExp(pattern);
    return true;
  } catch {
    return false;
  }
}

/**
 * Validates the firewall rule form, including the parts that depend on each other: a rule
 * filtered by hostname needs no tags, and one filtered by tags needs at least one.
 */
export const ruleSchema = z
  .object({
    priority: z.string(),
    action: z.enum(["allow", "deny"]),
    active: z.boolean(),
    sourceIpOption: z.enum(["all", "restrict"]),
    sourceIp: z.string(),
    usernameOption: z.enum(["all", "restrict"]),
    username: z.string(),
    filterOption: z.enum(["all", "hostname", "tags"]),
    hostname: z.string(),
    tags: z.array(z.string()),
  })
  .superRefine((values, ctx) => {
    if (values.priority.trim() === "") {
      ctx.addIssue({
        code: "custom",
        path: ["priority"],
        message: "Priority is required",
      });
    } else if (!Number.isInteger(Number(values.priority))) {
      ctx.addIssue({
        code: "custom",
        path: ["priority"],
        message: "Priority must be an integer",
      });
    } else if (Number(values.priority) <= 0) {
      ctx.addIssue({
        code: "custom",
        path: ["priority"],
        message: "Priority must be a positive integer",
      });
    }

    if (values.sourceIpOption === "restrict") {
      if (!values.sourceIp.trim()) {
        ctx.addIssue({
          code: "custom",
          path: ["sourceIp"],
          message: "Source IP pattern is required",
        });
      } else if (!isValidRegex(values.sourceIp)) {
        ctx.addIssue({
          code: "custom",
          path: ["sourceIp"],
          message: "Source IP must be a valid regular expression",
        });
      }
    }

    if (values.usernameOption === "restrict") {
      if (!values.username.trim()) {
        ctx.addIssue({
          code: "custom",
          path: ["username"],
          message: "Username pattern is required",
        });
      } else if (!isValidRegex(values.username)) {
        ctx.addIssue({
          code: "custom",
          path: ["username"],
          message: "Username must be a valid regular expression",
        });
      }
    }

    if (values.filterOption === "hostname") {
      if (!values.hostname.trim()) {
        ctx.addIssue({
          code: "custom",
          path: ["hostname"],
          message: "Hostname pattern is required",
        });
      } else if (!isValidRegex(values.hostname)) {
        ctx.addIssue({
          code: "custom",
          path: ["hostname"],
          message: "Hostname must be a valid regular expression",
        });
      }
    }

    if (values.filterOption === "tags") {
      if (values.tags.length === 0) {
        ctx.addIssue({
          code: "custom",
          path: ["tags"],
          message: "Select at least one tag",
        });
      } else if (values.tags.length > 3) {
        ctx.addIssue({
          code: "custom",
          path: ["tags"],
          message: "You can select up to 3 tags",
        });
      }
    }
  });

/**
 * The rule form's values, derived from the schema.
 */
export type RuleFormValues = z.infer<typeof ruleSchema>;

/**
 * What a new rule starts as: allow, active, and matching everything, so an unfinished rule is
 * visibly incomplete rather than quietly denying traffic.
 */
export const DEFAULT_VALUES: RuleFormValues = {
  priority: "",
  action: "allow",
  active: true,
  sourceIpOption: "all",
  sourceIp: "",
  usernameOption: "all",
  username: "",
  filterOption: "all",
  hostname: "",
  tags: [],
};

/**
 * Turns validated form values into the request body, sending only the filter branch that was
 * chosen — a hostname or a tag set, never both.
 */
export function buildRuleBody(values: RuleFormValues): FirewallRulesRequest {
  const filter: FirewallRulesRequest["filter"] =
    values.filterOption === "hostname"
      ? { hostname: values.hostname.trim() }
      : values.filterOption === "tags"
        ? { tags: values.tags }
        : { hostname: ".*" };

  return {
    priority: parseInt(values.priority, 10),
    action: values.action,
    active: values.active,
    source_ip:
      values.sourceIpOption === "restrict" ? values.sourceIp.trim() : ".*",
    username:
      values.usernameOption === "restrict" ? values.username.trim() : ".*",
    filter,
  };
}

/**
 * Fills the rule form from an existing rule. A hostname of ".*" is the API's way of saying
 * "any", so it is shown as an empty field rather than as a pattern the user did not write.
 */
export function buildRuleDefaults(rule: FirewallRulesResponse): RuleFormValues {
  const hasTags = rule.filter.tags.length > 0;
  const hasHostname =
    rule.filter.hostname !== undefined && rule.filter.hostname !== ".*";

  const filterOption: RuleFormValues["filterOption"] = hasTags
    ? "tags"
    : hasHostname
      ? "hostname"
      : "all";

  return {
    priority: String(rule.priority),
    action: rule.action,
    active: rule.active,
    sourceIpOption: rule.source_ip === ".*" ? "all" : "restrict",
    sourceIp: rule.source_ip === ".*" ? "" : rule.source_ip,
    usernameOption: rule.username === ".*" ? "all" : "restrict",
    username: rule.username === ".*" ? "" : rule.username,
    filterOption,
    hostname: filterOption === "hostname" ? (rule.filter.hostname ?? "") : "",
    tags: hasTags ? rule.filter.tags.map((t) => t.name) : [],
  };
}
