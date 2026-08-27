import { describe, it, expect } from "vitest";
import {
  sshIdentityStatus,
  sshIdentityEndOfLife,
  serviceAccountLifecyclePayload,
  sshIdentitySource,
  shortFingerprint,
  isAlreadyEnrolled,
} from "@/utils/sshIdentity";

const NOW = Date.UTC(2026, 0, 10);
const past = new Date(Date.UTC(2026, 0, 1)).toISOString();
const future = new Date(Date.UTC(2026, 0, 20)).toISOString();

describe("sshIdentityStatus", () => {
  it("returns null for a durable human identity", () => {
    expect(
      sshIdentityStatus(
        { expires_at: null, consumed_at: null, single_use: false },
        NOW,
      ),
    ).toBeNull();
  });

  it("marks a consumed key dead, even when also single-use", () => {
    expect(
      sshIdentityStatus(
        { expires_at: future, consumed_at: past, single_use: true },
        NOW,
      ),
    ).toMatchObject({ label: "Consumed", tone: "dead" });
  });

  it("marks an expired key dead", () => {
    expect(
      sshIdentityStatus(
        { expires_at: past, consumed_at: null, single_use: false },
        NOW,
      ),
    ).toMatchObject({ label: "Expired", tone: "dead" });
  });

  it("prefers single-use over an active TTL for an armed key", () => {
    expect(
      sshIdentityStatus(
        { expires_at: future, consumed_at: null, single_use: true },
        NOW,
      ),
    ).toMatchObject({ label: "Single-use", tone: "armed" });
  });

  it("names the deadline itself, not the distance to it", () => {
    const status = sshIdentityStatus({ expires_at: future, single_use: false }, NOW);
    expect(status?.label).toBe("Expires Jan 20, 2026");
    expect(status?.title).toBe("Jan 20, 2026, 00:00");
  });

  // A far-off expiry is a fact about the key; a near one is something to act on,
  // and only that one earns colour.
  it("colours the expiry only once the deadline is close", () => {
    const soon = new Date(Date.UTC(2026, 0, 15)).toISOString();
    const faraway = new Date(Date.UTC(2027, 0, 1)).toISOString();

    expect(sshIdentityStatus({ expires_at: soon, single_use: false }, NOW)?.tone).toBe("armed");
    expect(sshIdentityStatus({ expires_at: faraway, single_use: false }, NOW)?.tone).toBe("quiet");
  });
});

describe("serviceAccountLifecyclePayload", () => {
  it("omits expires_in when the selection is never (-1)", () => {
    expect(serviceAccountLifecyclePayload("-1", true)).toEqual({
      single_use: true,
    });
  });

  it("sends a positive selection as expires_in days", () => {
    expect(serviceAccountLifecyclePayload("30", false)).toEqual({
      single_use: false,
      expires_in: 30,
    });
  });
});

describe("sshIdentitySource", () => {
  it("names every origin", () => {
    expect(sshIdentitySource("browser").label).toBe("Browser");
    expect(sshIdentitySource("manual").label).toBe("Manual");
    expect(sshIdentitySource("approval").label).toBe("At login");
  });

  // Notable drives the one colour change in the column, so exactly one row in a
  // namespace can earn it: another browser's key is one you cannot reach either.
  it("marks only the browser in use as notable", () => {
    expect(sshIdentitySource("browser", true).notable).toBe(true);
    expect(sshIdentitySource("browser", false).notable).toBe(false);
    expect(sshIdentitySource("manual").notable).toBe(false);
    expect(sshIdentitySource("approval").notable).toBe(false);
  });

  it("singles out the browser the person is using now", () => {
    expect(sshIdentitySource("browser", true).label).toBe("This browser");
    expect(sshIdentitySource("browser", false).label).toBe("Browser");
  });

  // Revoking the key of the browser you are looking at costs you that terminal;
  // revoking another browser's costs you nothing here.
  it("says who pays for the revocation, in the right person", () => {
    expect(sshIdentitySource("browser", true).revokeConsequence).toMatch(
      /^This browser/,
    );
    expect(sshIdentitySource("browser", false).revokeConsequence).toMatch(
      /^That browser/,
    );
  });

  it("says revoking a browser key costs that browser its key, not an approval", () => {
    expect(sshIdentitySource("browser").revokeConsequence).not.toBe(
      sshIdentitySource("manual").revokeConsequence,
    );
  });

  // Only a browser key can be the current one; the flag must not leak elsewhere.
  it("ignores the current-browser flag for a key that is not a browser key", () => {
    expect(sshIdentitySource("manual", true)).toEqual(
      sshIdentitySource("manual"),
    );
  });

  // Rows written before the column existed, and any value a newer server sends,
  // must land on the copy that claims the least rather than blowing up.
  it("falls back to the manual copy for a missing or unknown source", () => {
    expect(sshIdentitySource(undefined)).toEqual(sshIdentitySource("manual"));
    expect(sshIdentitySource("something-new")).toEqual(
      sshIdentitySource("manual"),
    );
  });
});

describe("shortFingerprint", () => {
  it("keeps the algorithm prefix and enough body to tell two keys apart", () => {
    expect(
      shortFingerprint("SHA256:hHmU2OTPQjhAKm3ecpf4iw3lqWNCWaFbG1kBje0kn0"),
    ).toBe("SHA256:hHmU2OTPQjhA\u2026");
  });

  it("leaves a fingerprint that is already short alone", () => {
    expect(shortFingerprint("SHA256:abc123")).toBe("SHA256:abc123");
  });

  // Nothing in the product produces a bare fingerprint today, but truncating
  // one must not invent a prefix that was never there.
  it("handles a fingerprint with no prefix", () => {
    expect(shortFingerprint("hHmU2OTPQjhAKm3ecpf4")).toBe(
      "hHmU2OTPQjhA\u2026",
    );
  });
});

describe("sshIdentityEndOfLife", () => {
  // Every way a key can end lands in the same slot on screen, so they all have
  // to come back from the same place.
  it("phrases each ending for a column already headed Expires", () => {
    expect(sshIdentityEndOfLife({ single_use: false }, NOW)).toMatchObject({
      kind: "never",
      value: "never",
    });
    expect(
      sshIdentityEndOfLife({ expires_at: future, single_use: false }, NOW),
    ).toMatchObject({ kind: "expires", value: "Jan 20, 2026" });
    expect(
      sshIdentityEndOfLife({ expires_at: past, single_use: false }, NOW),
    ).toMatchObject({ kind: "expired", value: "expired" });
    expect(
      sshIdentityEndOfLife({ consumed_at: past, single_use: true }, NOW),
    ).toMatchObject({ kind: "consumed", value: "consumed" });
  });

  // A single-use key's deadline is not a date, and saying "on first use" in the
  // Expires slot is the honest way to put that.
  it("treats single-use as an ending, not as a missing date", () => {
    expect(sshIdentityEndOfLife({ single_use: true }, NOW)).toMatchObject({
      kind: "single-use",
      value: "on first use",
      tone: "armed",
    });
  });
});

describe("isAlreadyEnrolled", () => {
  it("recognizes the conflict the enroll endpoint answers with", () => {
    expect(isAlreadyEnrolled({ status: 409, headers: new Headers() })).toBe(
      true,
    );
  });

  it.each([400, 401, 403, 404, 500])(
    "does not treat %i as an existing enrollment",
    (status) => {
      expect(isAlreadyEnrolled({ status, headers: new Headers() })).toBe(false);
    },
  );

  it.each([
    ["a plain error", new Error("network down")],
    ["null", null],
    ["undefined", undefined],
    ["a bare object", {}],
  ])("does not treat %s as an existing enrollment", (_label, err) => {
    expect(isAlreadyEnrolled(err)).toBe(false);
  });
});
