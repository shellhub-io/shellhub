import { describe, it, expect } from "vitest";
import { validateNamespaceName } from "../validation";

describe("validateNamespaceName", () => {
  describe("valid names", () => {
    it.each([
      "abc",
      "a-b",
      "my-namespace",
      "a1b2c3",
      "1abc",
      "abc1",
      "a--b",
      "a".repeat(30),
    ])("accepts %p", (name) => {
      expect(validateNamespaceName(name)).toBeNull();
    });
  });

  describe("length rules", () => {
    it.each(["", "a", "ab"])("rejects %p as too short", (name) => {
      expect(validateNamespaceName(name)).toBe(
        "Name must be at least 3 characters",
      );
    });

    it("rejects names longer than 30 characters", () => {
      expect(validateNamespaceName("a".repeat(31))).toBe(
        "Name must be at most 30 characters",
      );
    });
  });

  describe("character rules", () => {
    it.each([
      "NameWithCaps",
      "has space",
      "has.dot",
      "has@at",
      "has#hash",
      "has$dollar",
      "has_underscore",
      "has/slash",
      "açento",
    ])("rejects %p", (name) => {
      expect(validateNamespaceName(name)).toBe(
        "Only lowercase letters, numbers, and hyphens (cannot start or end with hyphen)",
      );
    });
  });

  describe("structural rules", () => {
    it.each(["-leading", "trailing-", "-both-"])("rejects %p", (name) => {
      expect(validateNamespaceName(name)).toBe(
        "Only lowercase letters, numbers, and hyphens (cannot start or end with hyphen)",
      );
    });
  });
});
