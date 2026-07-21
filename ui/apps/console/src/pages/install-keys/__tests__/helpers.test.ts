import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { type InstallKey } from "@/client";
import {
  getExpiryInfo,
  getInstallKeyStatus,
  getKeyBlockers,
  getUsageInfo,
  installKeyDisplayName,
  isPairingKey,
  parseAllowedMacs,
  resolveEnrollmentSource,
  validateModeConfig,
  validateName,
} from "../helpers";

// The status/blocker/expiry helpers read Date.now(); pin it so "expired" is deterministic.
const NOW = new Date("2026-01-01T00:00:00Z");
const daysFromNow = (days: number) =>
  new Date(NOW.getTime() + days * 24 * 60 * 60 * 1000).toISOString();

function key(overrides: Partial<InstallKey>): InstallKey {
  return {
    revoked: false,
    disabled: false,
    expires_at: null,
    usage_limit: 0,
    used_times: 0,
    ...overrides,
  } as InstallKey;
}

beforeEach(() => {
  vi.useFakeTimers();
  vi.setSystemTime(NOW);
});

afterEach(() => {
  vi.useRealTimers();
});

describe("getInstallKeyStatus", () => {
  it("returns valid for a fresh, unlimited key", () => {
    expect(getInstallKeyStatus(key({}))).toBe("valid");
  });

  it("ranks revoked above every other blocker", () => {
    const k = key({
      revoked: true,
      disabled: true,
      expires_at: daysFromNow(-1),
      usage_limit: 1,
      used_times: 5,
    });
    expect(getInstallKeyStatus(k)).toBe("revoked");
  });

  it("ranks disabled above expiry and overuse", () => {
    const k = key({
      disabled: true,
      expires_at: daysFromNow(-1),
      usage_limit: 1,
      used_times: 1,
    });
    expect(getInstallKeyStatus(k)).toBe("disabled");
  });

  it("ranks an elapsed expiry above overuse", () => {
    expect(
      getInstallKeyStatus(
        key({ expires_at: daysFromNow(-1), usage_limit: 1, used_times: 1 }),
      ),
    ).toBe("expired");
  });

  it("reports overused only when a positive limit is reached", () => {
    expect(getInstallKeyStatus(key({ usage_limit: 3, used_times: 3 }))).toBe(
      "overused",
    );
    expect(getInstallKeyStatus(key({ usage_limit: 3, used_times: 2 }))).toBe(
      "valid",
    );
    // usage_limit 0 means unlimited, so it never counts as overused however high used_times climbs.
    expect(getInstallKeyStatus(key({ usage_limit: 0, used_times: 999 }))).toBe(
      "valid",
    );
  });
});

describe("getKeyBlockers", () => {
  it("surfaces each blocker independently and flags inert when any is set", () => {
    const blockers = getKeyBlockers(
      key({ expires_at: daysFromNow(-1), usage_limit: 2, used_times: 2 }),
    );
    expect(blockers).toMatchObject({
      revoked: false,
      disabled: false,
      expired: true,
      overused: true,
      inert: true,
    });
  });

  it("is not inert for a live key", () => {
    expect(getKeyBlockers(key({})).inert).toBe(false);
  });
});

describe("getUsageInfo", () => {
  it("treats usage_limit 0 as unlimited", () => {
    expect(getUsageInfo(key({ usage_limit: 0, used_times: 12 }))).toMatchObject(
      {
        kind: "unlimited",
        limit: 0,
        ratio: 0,
        exhausted: false,
      },
    );
  });

  it("treats usage_limit 1 as single-use", () => {
    expect(getUsageInfo(key({ usage_limit: 1, used_times: 0 })).kind).toBe(
      "single",
    );
  });

  it("treats a limit of two or more as limited and computes the fill ratio", () => {
    const info = getUsageInfo(key({ usage_limit: 4, used_times: 1 }));
    expect(info.kind).toBe("limited");
    expect(info.ratio).toBe(0.25);
    expect(info.exhausted).toBe(false);
  });

  it("clamps the ratio at 1 and marks exhausted at the limit", () => {
    const info = getUsageInfo(key({ usage_limit: 2, used_times: 5 }));
    expect(info.ratio).toBe(1);
    expect(info.exhausted).toBe(true);
  });
});

describe("getExpiryInfo", () => {
  it("reads a null expiry as a muted Never", () => {
    expect(getExpiryInfo(null)).toEqual({ label: "Never", tone: "muted" });
    expect(getExpiryInfo(undefined)).toEqual({ label: "Never", tone: "muted" });
  });

  it("is danger once elapsed", () => {
    expect(getExpiryInfo(daysFromNow(-1)).tone).toBe("danger");
  });

  it("warns within the last seven calendar days and is normal beyond", () => {
    expect(getExpiryInfo(daysFromNow(3)).tone).toBe("warning");
    expect(getExpiryInfo(daysFromNow(7)).tone).toBe("warning");
    expect(getExpiryInfo(daysFromNow(30)).tone).toBe("normal");
  });

  it("labels a live expiry with its date", () => {
    expect(getExpiryInfo(daysFromNow(30)).label).toBeTruthy();
  });
});

describe("validateModeConfig", () => {
  it("accepts a non-webhook, non-allowlist mode with no extra config", () => {
    expect(validateModeConfig("automatic", "", "", [])).toBe("");
  });

  it("requires an http(s) url and a secret for webhook mode", () => {
    expect(validateModeConfig("webhook", "not-a-url", "s", [])).not.toBe("");
    expect(
      validateModeConfig("webhook", "https://hook.example", "", []),
    ).not.toBe("");
    expect(validateModeConfig("webhook", "https://hook.example", "s", [])).toBe(
      "",
    );
  });

  it("requires at least one MAC for allowlist mode", () => {
    expect(validateModeConfig("allowlist", "", "", [])).not.toBe("");
    expect(validateModeConfig("allowlist", "", "", ["aa:bb:cc:dd:ee:ff"])).toBe(
      "",
    );
  });

  it("lets a blank secret pass when it is optional (editing a webhook key)", () => {
    expect(
      validateModeConfig("webhook", "https://hook.example", "", []),
    ).not.toBe("");
    expect(
      validateModeConfig("webhook", "https://hook.example", "", [], {
        secretOptional: true,
      }),
    ).toBe("");
    // The URL is still validated even when the secret is optional.
    expect(
      validateModeConfig("webhook", "bad", "", [], { secretOptional: true }),
    ).not.toBe("");
  });
});

describe("validateName", () => {
  it("accepts a valid name", () => {
    expect(validateName("ci-runners")).toBe("");
    expect(validateName("edge_01")).toBe("");
  });

  it("rejects too short, too long, and illegal characters", () => {
    expect(validateName("ab")).not.toBe("");
    expect(validateName("a".repeat(21))).not.toBe("");
    expect(validateName("has space")).not.toBe("");
    expect(validateName("bang!")).not.toBe("");
  });
});

describe("parseAllowedMacs", () => {
  it("lowercases, trims, drops blanks, and dedupes", () => {
    expect(
      parseAllowedMacs(
        "AA:BB:CC:DD:EE:FF\n  aa:bb:cc:dd:ee:ff \n\n11:22:33:44:55:66",
      ),
    ).toEqual(["aa:bb:cc:dd:ee:ff", "11:22:33:44:55:66"]);
  });

  it("returns an empty list for blank input", () => {
    expect(parseAllowedMacs("\n  \n")).toEqual([]);
  });
});

describe("enrollment source", () => {
  const legacy = key({ id: "leg", name: "legacy", type: "legacy" });
  const pairing = key({ id: "pair", name: "pairing", type: "pairing" });
  const real = key({ id: "abc", name: "fleet", type: "user" });

  it("resolves a device to its enrollment source, or null when unmatched", () => {
    const keys = [legacy, pairing, real];
    expect(resolveEnrollmentSource(undefined, keys)).toBeNull();
    expect(resolveEnrollmentSource("nope", keys)).toBeNull();
    expect(resolveEnrollmentSource("abc", keys)).toEqual({
      kind: "key",
      name: "fleet",
    });
    expect(resolveEnrollmentSource("leg", keys)).toEqual({ kind: "legacy" });
    expect(resolveEnrollmentSource("pair", keys)).toEqual({ kind: "pairing" });
  });

  it("labels the pairing system key apart from the legacy one", () => {
    expect(isPairingKey(pairing)).toBe(true);
    expect(isPairingKey(legacy)).toBe(false);
    expect(isPairingKey(real)).toBe(false);

    expect(installKeyDisplayName(pairing)).toBe("Pairing code");
    expect(installKeyDisplayName(legacy)).toBe("Tenant-only registration");
    expect(installKeyDisplayName(real)).toBe("fleet");
  });
});
