import { readFileSync } from "node:fs";
import { join } from "node:path";
import { describe, expect, it } from "vitest";
import { routes } from "@/lib/routes";

interface Rule {
  from: string;
  to: string;
  status: string;
}

const readRules = (): Rule[] =>
  readFileSync(join(import.meta.dirname, "../../../public/_redirects"), "utf-8")
    .split("\n")
    .map((line) => line.trim())
    .filter((line) => line !== "" && !line.startsWith("#"))
    .map((line) => {
      const [from, to, status] = line.split(/\s+/);

      return { from, to, status };
    });

const firstRuleThatMatches = (rules: Rule[], path: string): Rule | undefined =>
  rules.find((rule) =>
    rule.from.endsWith("*")
      ? path.startsWith(rule.from.slice(0, -1))
      : rule.from === path,
  );

const addressOf = (rule: Rule): string =>
  rule.from.endsWith("/*") ? rule.from.slice(0, -2) : rule.from;

describe("_redirects", () => {
  const rules = readRules();

  it("carries a rule for every address the old site published", () => {
    expect(rules.length).toBeGreaterThan(150);
  });

  it("answers every rule with a permanent redirect", () => {
    for (const { from, status } of rules) {
      expect(status, `"${from}" is not a 301`).toBe("301");
    }
  });

  it("keeps every destination on this site", () => {
    for (const { from, to } of rules) {
      expect(to.startsWith("/"), `"${from}" leaves the site`).toBe(true);
    }
  });

  it("sends an address and its trailing-slash form to the same page", () => {
    for (const rule of rules) {
      const bare = addressOf(rule);
      const forms = rule.from.endsWith("*") ? [bare, bare + "/"] : [
        bare.endsWith("/") ? bare.slice(0, -1) : bare + "/",
      ];

      for (const form of forms) {
        expect(
          firstRuleThatMatches(rules, form)?.to,
          `"${rule.from}" redirects, but "${form}" does not reach the same page`,
        ).toBe(rule.to);
      }
    }
  });

  it("redirects to pages that exist", () => {
    const pages = routes();

    for (const { from, to } of rules) {
      const page = to.split("#")[0];

      expect(pages.has(page), `"${from}" points at "${page}"`).toBe(true);
    }
  });

  it("leaves every rule reachable, none shadowed by an earlier one", () => {
    for (const rule of rules) {
      const winner = firstRuleThatMatches(rules, rule.from);

      expect(
        winner?.to,
        `"${rule.from}" never applies: "${winner?.from}" matches first`,
      ).toBe(rule.to);
    }
  });

  it("redirects no address this site serves itself", () => {
    for (const route of routes()) {
      const rule = firstRuleThatMatches(rules, route);

      expect(
        rule,
        `"${route}" is a page here, and "${rule?.from}" redirects it away`,
      ).toBeUndefined();
    }
  });
});
