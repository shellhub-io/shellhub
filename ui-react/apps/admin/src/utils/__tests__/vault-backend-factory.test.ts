import { describe, it, expect, beforeEach } from "vitest";
import { getVaultBackend, resetVaultBackend } from "../vault-backend-factory";
import { LocalVaultBackend } from "../vault-backend-local";

beforeEach(() => {
  resetVaultBackend();
});

describe("getVaultBackend", () => {
  it("returns a LocalVaultBackend instance", () => {
    expect(getVaultBackend()).toBeInstanceOf(LocalVaultBackend);
  });

  it("returns the same instance on repeated calls (singleton)", () => {
    const first = getVaultBackend();
    const second = getVaultBackend();
    expect(first).toBe(second);
  });

  it("returns a new instance after resetVaultBackend()", () => {
    const first = getVaultBackend();
    resetVaultBackend();
    const second = getVaultBackend();
    expect(first).not.toBe(second);
  });

  it("returned backend satisfies the IVaultBackend interface", () => {
    const backend = getVaultBackend();
    expect(typeof backend.loadMeta).toBe("function");
    expect(typeof backend.saveMeta).toBe("function");
    expect(typeof backend.loadData).toBe("function");
    expect(typeof backend.saveData).toBe("function");
    expect(typeof backend.clear).toBe("function");
    expect(typeof backend.loadLegacyKeys).toBe("function");
    expect(typeof backend.clearLegacyKeys).toBe("function");
  });
});

describe("resetVaultBackend", () => {
  it("is idempotent when called without a prior getVaultBackend call", () => {
    expect(() => {
      resetVaultBackend();
      resetVaultBackend();
    }).not.toThrow();
  });
});
