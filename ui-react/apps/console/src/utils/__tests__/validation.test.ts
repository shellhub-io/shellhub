import { describe, it, expect } from "vitest";
import {
  validateEmail,
  validateName,
  validateNamespaceName,
  validatePassword,
  validateUsername,
} from "../validation";

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

describe("validateName", () => {
  it.each(["a", "Alice Smith", "X", "a".repeat(64), "  Alice  "])(
    "accepts %p",
    (value) => {
      expect(validateName(value)).toBeNull();
    },
  );

  it.each(["", "   "])("rejects empty/whitespace %p as required", (value) => {
    expect(validateName(value)).toBe("Name is required");
  });

  it("rejects names longer than 64 characters", () => {
    expect(validateName("a".repeat(65))).toBe(
      "Name must be at most 64 characters",
    );
  });
});

describe("validateUsername", () => {
  it.each([
    "alice",
    "alice123",
    "alice.smith",
    "alice_smith",
    "alice-smith",
    "user@example",
    "abc",
    "a".repeat(32),
  ])("accepts %p", (value) => {
    expect(validateUsername(value)).toBeNull();
  });

  it.each(["", "   "])("rejects empty/whitespace %p as required", (value) => {
    expect(validateUsername(value)).toBe("Username is required");
  });

  it.each([
    "ab", // too short
    "a".repeat(33), // too long
    "Alice", // uppercase
    "ALICE",
    "has space",
    "has#hash",
    "has/slash",
    "has!bang",
  ])("rejects %p with the hint message", (value) => {
    expect(validateUsername(value)).toMatch(/^3-32 characters/);
  });
});

describe("validateEmail", () => {
  it.each([
    "alice@example.com",
    "user.name@example.co.uk",
    "user+tag@example.org",
  ])("accepts %p", (value) => {
    expect(validateEmail(value)).toBeNull();
  });

  it.each(["", "   "])("rejects empty/whitespace %p as required", (value) => {
    expect(validateEmail(value)).toBe("Email is required");
  });

  it.each(["plainstring", "no@domain", "missing.at.sign", "two@@signs.com"])(
    "rejects %p",
    (value) => {
      expect(validateEmail(value)).toBe("Enter a valid email address");
    },
  );
});

describe("validatePassword", () => {
  it.each(["12345", "a".repeat(32), "P@ssw0rd!"])("accepts %p", (value) => {
    expect(validatePassword(value)).toBeNull();
  });

  it.each(["", "1234", "a".repeat(33)])("rejects %p", (value) => {
    expect(validatePassword(value)).toBe(
      "Password must be 5–32 characters long",
    );
  });
});
