import { describe, it, expect, beforeEach } from "vitest";
import { getVaultBackend } from "../vault-backend-factory";
import { LocalVaultBackend } from "../vault-backend-local";

beforeEach(() => {
  localStorage.clear();
});

describe("getVaultBackend", () => {
  it("returns a LocalVaultBackend instance", () => {
    expect(getVaultBackend()).toBeInstanceOf(LocalVaultBackend);
  });

  it("returns a fresh instance on each call (no singleton)", () => {
    const first = getVaultBackend();
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

  it("scopes storage keys when scope is provided", () => {
    const backend = getVaultBackend({ user: "testuser", tenant: "test-tenant" });
    const meta = {
      version: 1 as const,
      salt: "c2FsdA==",
      iterations: 600000,
      verifier: "dmVyaWZpZXI=",
      verifierIv: "aXY=",
    };
    backend.saveMeta(meta);
    expect(localStorage.getItem("shellhub-vault-meta:testuser:test-tenant")).toBe(JSON.stringify(meta));
    expect(localStorage.getItem("shellhub-vault-meta")).toBeNull();
  });

  it("uses unscoped keys when no scope is provided", () => {
    const backend = getVaultBackend();
    const meta = {
      version: 1 as const,
      salt: "c2FsdA==",
      iterations: 600000,
      verifier: "dmVyaWZpZXI=",
      verifierIv: "aXY=",
    };
    backend.saveMeta(meta);
    expect(localStorage.getItem("shellhub-vault-meta")).toBe(JSON.stringify(meta));
  });
});
