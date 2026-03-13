import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { hasSeenWelcome, markWelcomeSeen } from "../welcomeState";

beforeEach(() => {
  localStorage.clear();
});

afterEach(() => {
  localStorage.clear();
});

describe("hasSeenWelcome", () => {
  it("returns false when localStorage is empty", () => {
    expect(hasSeenWelcome("tenant-abc")).toBe(false);
  });

  it("returns false when a different tenant is marked", () => {
    localStorage.setItem("shellhub:welcomed:tenant-other", "true");
    expect(hasSeenWelcome("tenant-abc")).toBe(false);
  });

  it("returns true when the tenant key is set to 'true'", () => {
    localStorage.setItem("shellhub:welcomed:tenant-abc", "true");
    expect(hasSeenWelcome("tenant-abc")).toBe(true);
  });

  it("returns false when the tenant key holds an unexpected value", () => {
    localStorage.setItem("shellhub:welcomed:tenant-abc", "yes");
    expect(hasSeenWelcome("tenant-abc")).toBe(false);
  });
});

describe("markWelcomeSeen", () => {
  it("sets the per-tenant key to 'true'", () => {
    markWelcomeSeen("tenant-abc");
    expect(localStorage.getItem("shellhub:welcomed:tenant-abc")).toBe("true");
  });

  it("makes hasSeenWelcome return true for the marked tenant", () => {
    markWelcomeSeen("tenant-abc");
    expect(hasSeenWelcome("tenant-abc")).toBe(true);
  });

  it("does not affect other tenants", () => {
    markWelcomeSeen("tenant-abc");
    expect(hasSeenWelcome("tenant-other")).toBe(false);
  });

  it("is idempotent — calling twice leaves a single key", () => {
    markWelcomeSeen("tenant-abc");
    markWelcomeSeen("tenant-abc");
    expect(localStorage.getItem("shellhub:welcomed:tenant-abc")).toBe("true");
  });

  it("marks multiple tenants independently", () => {
    markWelcomeSeen("tenant-abc");
    markWelcomeSeen("tenant-xyz");
    expect(hasSeenWelcome("tenant-abc")).toBe(true);
    expect(hasSeenWelcome("tenant-xyz")).toBe(true);
    expect(hasSeenWelcome("tenant-other")).toBe(false);
  });
});
