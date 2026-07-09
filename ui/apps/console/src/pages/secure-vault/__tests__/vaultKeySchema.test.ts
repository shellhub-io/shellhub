import { describe, it, expect, vi, beforeEach } from "vitest";
import { validatePrivateKey } from "@/utils/sshKeys";
import type { VaultKeyEntry } from "@/types/vault";
import {
  vaultKeySchema,
  buildVaultKeyDefaults,
  buildVaultKeyPayload,
  type VaultKeyFormValues,
} from "../vaultKeySchema";

vi.mock("@/utils/sshKeys", () => ({
  validatePrivateKey: vi.fn(),
}));

function makeValues(
  overrides: Partial<VaultKeyFormValues> = {},
): VaultKeyFormValues {
  return {
    name: "My Key",
    data: "-----BEGIN OPENSSH PRIVATE KEY-----\nvalid\n-----END OPENSSH PRIVATE KEY-----",
    encrypted: false,
    passphrase: "",
    ...overrides,
  };
}

/** First validation message per field. */
function errorsFor(
  values: VaultKeyFormValues,
): Partial<Record<keyof VaultKeyFormValues, string>> {
  const result = vaultKeySchema.safeParse(values);
  if (result.success) return {};

  const errors: Partial<Record<keyof VaultKeyFormValues, string>> = {};
  for (const issue of result.error.issues) {
    const key = issue.path[0] as keyof VaultKeyFormValues;
    if (key && errors[key] === undefined) errors[key] = issue.message;
  }
  return errors;
}

describe("vaultKeySchema", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(validatePrivateKey).mockReturnValue({
      valid: true,
      encrypted: false,
    });
  });

  describe("name", () => {
    it("rejects an empty name", () => {
      expect(errorsFor(makeValues({ name: "" })).name).toBeDefined();
    });

    it("rejects a whitespace-only name", () => {
      expect(errorsFor(makeValues({ name: "   " })).name).toBeDefined();
    });

    it("accepts a non-empty name", () => {
      expect(errorsFor(makeValues({ name: "Prod" })).name).toBeUndefined();
    });
  });

  describe("data", () => {
    it("rejects an empty key", () => {
      expect(errorsFor(makeValues({ data: "" })).data).toBeDefined();
    });

    it("rejects a whitespace-only key", () => {
      expect(errorsFor(makeValues({ data: "   " })).data).toBeDefined();
    });

    it("surfaces the validator's error message for an invalid key", () => {
      vi.mocked(validatePrivateKey).mockReturnValue({
        valid: false,
        error: "Invalid private key format.",
      });
      expect(errorsFor(makeValues()).data).toBe("Invalid private key format.");
    });

    it("accepts a valid key", () => {
      expect(errorsFor(makeValues()).data).toBeUndefined();
    });
  });

  describe("passphrase", () => {
    it("requires a passphrase when the key is encrypted", () => {
      expect(
        errorsFor(makeValues({ encrypted: true, passphrase: "" })).passphrase,
      ).toBeDefined();
    });

    it("rejects a whitespace-only passphrase when encrypted", () => {
      expect(
        errorsFor(makeValues({ encrypted: true, passphrase: "   " }))
          .passphrase,
      ).toBeDefined();
    });

    it("accepts a passphrase when encrypted", () => {
      expect(
        errorsFor(makeValues({ encrypted: true, passphrase: "secret" }))
          .passphrase,
      ).toBeUndefined();
    });

    it("ignores the passphrase when the key is not encrypted", () => {
      expect(
        errorsFor(makeValues({ encrypted: false, passphrase: "" })).passphrase,
      ).toBeUndefined();
    });
  });
});

describe("buildVaultKeyDefaults", () => {
  it("returns empty defaults for a null entry", () => {
    expect(buildVaultKeyDefaults(null)).toEqual({
      name: "",
      data: "",
      encrypted: false,
      passphrase: "",
    });
  });

  it("maps an entry, deriving 'encrypted' from hasPassphrase and clearing the passphrase", () => {
    const entry: VaultKeyEntry = {
      id: "k1",
      name: "Prod",
      data: "PEMDATA",
      hasPassphrase: true,
      fingerprint: "aa:bb",
      createdAt: "2024-01-01T00:00:00Z",
      updatedAt: "2024-01-01T00:00:00Z",
    };
    expect(buildVaultKeyDefaults(entry)).toEqual({
      name: "Prod",
      data: "PEMDATA",
      encrypted: true,
      passphrase: "",
    });
  });
});

describe("buildVaultKeyPayload", () => {
  it("trims name and data and passes through fingerprint/algorithm", () => {
    const payload = buildVaultKeyPayload(
      makeValues({ name: "  Prod  ", data: "  PEM  ", encrypted: true }),
      "aa:bb:cc",
      "Ed25519",
    );
    expect(payload).toEqual({
      name: "Prod",
      data: "PEM",
      hasPassphrase: true,
      fingerprint: "aa:bb:cc",
      algorithm: "Ed25519",
    });
  });
});
