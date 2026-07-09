import { describe, it, expect } from "vitest";
import {
  ruleSchema,
  buildRuleBody,
  buildRuleDefaults,
  type RuleFormValues,
} from "../ruleSchema";
import type { FirewallRulesResponse } from "@/client";

// ── Factories ─────────────────────────────────────────────────────────────────

function makeValues(overrides: Partial<RuleFormValues> = {}): RuleFormValues {
  return {
    priority: "10",
    action: "allow",
    active: true,
    sourceIpOption: "all",
    sourceIp: "",
    usernameOption: "all",
    username: "",
    filterOption: "all",
    hostname: "",
    tags: [],
    ...overrides,
  };
}

/** First validation message per field, mirroring the RHF resolver shape. */
function errorsFor(
  values: RuleFormValues,
): Partial<Record<keyof RuleFormValues, string>> {
  const result = ruleSchema.safeParse(values);
  if (result.success) return {};

  const errors: Partial<Record<keyof RuleFormValues, string>> = {};
  for (const issue of result.error.issues) {
    const key = issue.path[0] as keyof RuleFormValues;
    if (key && errors[key] === undefined) errors[key] = issue.message;
  }
  return errors;
}

function makeResponse(
  overrides: Partial<FirewallRulesResponse> = {},
): FirewallRulesResponse {
  return {
    id: "rule-1",
    tenant_id: "tenant-abc",
    priority: 42,
    action: "allow",
    active: true,
    source_ip: ".*",
    username: ".*",
    filter: { hostname: ".*", tags: [] },
    ...overrides,
  };
}

// ── priority field ────────────────────────────────────────────────────────────

describe("ruleSchema — priority", () => {
  it("rejects an empty priority", () => {
    expect(errorsFor(makeValues({ priority: "" })).priority).toBeDefined();
  });

  it("rejects '0' as a non-positive priority", () => {
    expect(errorsFor(makeValues({ priority: "0" })).priority).toBeDefined();
  });

  it("rejects a negative priority", () => {
    expect(errorsFor(makeValues({ priority: "-5" })).priority).toBeDefined();
  });

  it("rejects a non-integer priority", () => {
    expect(errorsFor(makeValues({ priority: "3.5" })).priority).toBeDefined();
  });

  it("rejects a non-numeric string as priority", () => {
    expect(errorsFor(makeValues({ priority: "abc" })).priority).toBeDefined();
  });

  it("accepts positive integer priorities", () => {
    expect(errorsFor(makeValues({ priority: "100" })).priority).toBeUndefined();
  });
});

// ── source_ip field (restrict mode) ──────────────────────────────────────────

describe("ruleSchema — sourceIp", () => {
  it("requires a non-empty sourceIp in restrict mode", () => {
    expect(
      errorsFor(makeValues({ sourceIpOption: "restrict", sourceIp: "" })).sourceIp,
    ).toBeDefined();
  });

  it("rejects an invalid regex as sourceIp in restrict mode", () => {
    expect(
      errorsFor(makeValues({ sourceIpOption: "restrict", sourceIp: "[invalid" }))
        .sourceIp,
    ).toBeDefined();
  });

  it("accepts a valid regex as sourceIp in restrict mode", () => {
    expect(
      errorsFor(
        makeValues({ sourceIpOption: "restrict", sourceIp: "192\\.168\\..*" }),
      ).sourceIp,
    ).toBeUndefined();
  });

  it("ignores sourceIp validation when sourceIpOption is 'all'", () => {
    expect(
      errorsFor(makeValues({ sourceIpOption: "all", sourceIp: "" })).sourceIp,
    ).toBeUndefined();
  });
});

// ── username field (restrict mode) ───────────────────────────────────────────

describe("ruleSchema — username", () => {
  it("requires a non-empty username in restrict mode", () => {
    expect(
      errorsFor(makeValues({ usernameOption: "restrict", username: "" })).username,
    ).toBeDefined();
  });

  it("rejects an invalid regex as username in restrict mode", () => {
    expect(
      errorsFor(makeValues({ usernameOption: "restrict", username: "(unclosed" }))
        .username,
    ).toBeDefined();
  });

  it("accepts a valid regex as username in restrict mode", () => {
    expect(
      errorsFor(makeValues({ usernameOption: "restrict", username: "root|admin" }))
        .username,
    ).toBeUndefined();
  });

  it("ignores username validation when usernameOption is 'all'", () => {
    expect(
      errorsFor(makeValues({ usernameOption: "all", username: "" })).username,
    ).toBeUndefined();
  });
});

// ── hostname field (hostname mode) ───────────────────────────────────────────

describe("ruleSchema — hostname", () => {
  it("requires a non-empty hostname when filterOption is 'hostname'", () => {
    expect(
      errorsFor(makeValues({ filterOption: "hostname", hostname: "" })).hostname,
    ).toBeDefined();
  });

  it("rejects an invalid regex as hostname", () => {
    expect(
      errorsFor(makeValues({ filterOption: "hostname", hostname: "**bad" }))
        .hostname,
    ).toBeDefined();
  });

  it("accepts a valid regex as hostname", () => {
    expect(
      errorsFor(makeValues({ filterOption: "hostname", hostname: "web-.*" }))
        .hostname,
    ).toBeUndefined();
  });

  it("ignores hostname validation when filterOption is not 'hostname'", () => {
    expect(
      errorsFor(makeValues({ filterOption: "all", hostname: "" })).hostname,
    ).toBeUndefined();
  });
});

// ── tags field (tags mode) ────────────────────────────────────────────────────

describe("ruleSchema — tags", () => {
  it("rejects 0 tags in tags mode", () => {
    expect(errorsFor(makeValues({ filterOption: "tags", tags: [] })).tags).toBeDefined();
  });

  it("rejects more than 3 tags in tags mode", () => {
    expect(
      errorsFor(makeValues({ filterOption: "tags", tags: ["t1", "t2", "t3", "t4"] }))
        .tags,
    ).toBeDefined();
  });

  it.each([
    { count: 1, tags: ["t1"] },
    { count: 2, tags: ["t1", "t2"] },
    { count: 3, tags: ["t1", "t2", "t3"] },
  ])("accepts $count tag(s) in tags mode", ({ tags }) => {
    expect(errorsFor(makeValues({ filterOption: "tags", tags })).tags).toBeUndefined();
  });

  it("ignores tag count validation when filterOption is not 'tags'", () => {
    expect(errorsFor(makeValues({ filterOption: "all", tags: [] })).tags).toBeUndefined();
  });
});

// ── buildRuleBody ─────────────────────────────────────────────────────────────

describe("buildRuleBody", () => {
  it("builds a body with source_ip='.*' and username='.*' when both are 'all'", () => {
    const body = buildRuleBody(makeValues({ priority: "5" }));
    expect(body.source_ip).toBe(".*");
    expect(body.username).toBe(".*");
    expect(body.priority).toBe(5);
  });

  it("uses the restrict sourceIp value when sourceIpOption is 'restrict'", () => {
    const body = buildRuleBody(
      makeValues({ sourceIpOption: "restrict", sourceIp: "10\\.0\\..*" }),
    );
    expect(body.source_ip).toBe("10\\.0\\..*");
  });

  it("uses the restrict username value when usernameOption is 'restrict'", () => {
    const body = buildRuleBody(
      makeValues({ usernameOption: "restrict", username: "root" }),
    );
    expect(body.username).toBe("root");
  });

  it("builds filter with hostname when filterOption is 'hostname'", () => {
    const body = buildRuleBody(
      makeValues({ filterOption: "hostname", hostname: "web-.*" }),
    );
    expect(body.filter).toEqual({ hostname: "web-.*" });
  });

  it("builds filter with tags when filterOption is 'tags'", () => {
    const body = buildRuleBody(
      makeValues({ filterOption: "tags", tags: ["prod", "db"] }),
    );
    expect(body.filter).toEqual({ tags: ["prod", "db"] });
  });

  it("builds filter with hostname='.*' when filterOption is 'all'", () => {
    const body = buildRuleBody(makeValues({ filterOption: "all" }));
    expect(body.filter).toEqual({ hostname: ".*" });
  });

  it("maps action and active from form values", () => {
    const body = buildRuleBody(makeValues({ action: "deny", active: false }));
    expect(body.action).toBe("deny");
    expect(body.active).toBe(false);
  });
});

// ── buildRuleDefaults ─────────────────────────────────────────────────────────

describe("buildRuleDefaults", () => {
  it("round-trips a rule with source_ip='.*' and username='.*'", () => {
    const defaults = buildRuleDefaults(makeResponse());
    expect(defaults.priority).toBe("42");
    expect(defaults.sourceIpOption).toBe("all");
    expect(defaults.sourceIp).toBe("");
    expect(defaults.usernameOption).toBe("all");
    expect(defaults.username).toBe("");
    expect(defaults.filterOption).toBe("all");
  });

  it("round-trips a rule with a restricted source_ip", () => {
    const defaults = buildRuleDefaults(
      makeResponse({ source_ip: "192\\.168\\..*" }),
    );
    expect(defaults.sourceIpOption).toBe("restrict");
    expect(defaults.sourceIp).toBe("192\\.168\\..*");
  });

  it("round-trips a rule with a restricted username", () => {
    const defaults = buildRuleDefaults(makeResponse({ username: "root" }));
    expect(defaults.usernameOption).toBe("restrict");
    expect(defaults.username).toBe("root");
  });

  it("round-trips a rule with a hostname filter (non '.*')", () => {
    const defaults = buildRuleDefaults(
      makeResponse({ filter: { hostname: "web-.*", tags: [] } }),
    );
    expect(defaults.filterOption).toBe("hostname");
    expect(defaults.hostname).toBe("web-.*");
    expect(defaults.tags).toEqual([]);
  });

  it("round-trips a rule with a tags filter", () => {
    const defaults = buildRuleDefaults(
      makeResponse({
        filter: {
          hostname: undefined,
          tags: [
            { name: "prod", tenant_id: "t", created_at: "2024-01-01", updated_at: "2024-01-01" },
            { name: "db", tenant_id: "t", created_at: "2024-01-01", updated_at: "2024-01-01" },
          ],
        },
      }),
    );
    expect(defaults.filterOption).toBe("tags");
    expect(defaults.tags).toEqual(["prod", "db"]);
    expect(defaults.hostname).toBe("");
  });

  it("treats filter.tags present without a hostname key as tags mode", () => {
    const defaults = buildRuleDefaults(
      makeResponse({
        filter: {
          hostname: undefined,
          tags: [{ name: "edge", tenant_id: "t", created_at: "2024-01-01", updated_at: "2024-01-01" }],
        },
      }),
    );
    expect(defaults.filterOption).toBe("tags");
  });

  it("treats filter.hostname='.*' with empty tags as 'all' mode", () => {
    const defaults = buildRuleDefaults(
      makeResponse({ filter: { hostname: ".*", tags: [] } }),
    );
    expect(defaults.filterOption).toBe("all");
    expect(defaults.hostname).toBe("");
  });
});
