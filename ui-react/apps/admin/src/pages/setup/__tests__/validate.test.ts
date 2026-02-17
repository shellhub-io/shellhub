import { describe, it, expect } from "vitest";
import { validate } from "../validate";

const valid = {
  name: "Admin User",
  username: "admin",
  email: "admin@example.com",
  password: "secret123",
  confirmPassword: "secret123",
};

describe("Setup validate", () => {
  it("returns no errors for valid input", () => {
    expect(validate(valid)).toEqual({});
  });

  describe("name", () => {
    it("rejects empty name", () => {
      expect(validate({ ...valid, name: "" }).name).toBeDefined();
    });

    it("accepts 1-char name", () => {
      expect(validate({ ...valid, name: "A" }).name).toBeUndefined();
    });

    it("accepts 64-char name", () => {
      expect(validate({ ...valid, name: "a".repeat(64) }).name).toBeUndefined();
    });

    it("rejects name over 64 chars", () => {
      expect(validate({ ...valid, name: "a".repeat(65) }).name).toBeDefined();
    });
  });

  describe("username", () => {
    it("rejects shorter than 3 chars", () => {
      expect(validate({ ...valid, username: "ab" }).username).toBeDefined();
    });

    it("accepts 3-char username", () => {
      expect(validate({ ...valid, username: "abc" }).username).toBeUndefined();
    });

    it("accepts 32-char username", () => {
      expect(
        validate({ ...valid, username: "a".repeat(32) }).username,
      ).toBeUndefined();
    });

    it("rejects longer than 32 chars", () => {
      expect(
        validate({ ...valid, username: "a".repeat(33) }).username,
      ).toBeDefined();
    });

    it("accepts lowercase, numbers, hyphens, dots, underscores, @", () => {
      expect(
        validate({ ...valid, username: "user-name_1.0@test" }).username,
      ).toBeUndefined();
    });

    it("rejects uppercase letters", () => {
      expect(validate({ ...valid, username: "Admin" }).username).toBeDefined();
    });

    it("rejects spaces", () => {
      expect(
        validate({ ...valid, username: "my user" }).username,
      ).toBeDefined();
    });

    it("rejects special characters", () => {
      expect(validate({ ...valid, username: "user!" }).username).toBeDefined();
    });
  });

  describe("email", () => {
    it("rejects empty email", () => {
      expect(validate({ ...valid, email: "" }).email).toBeDefined();
    });

    it("rejects missing @", () => {
      expect(validate({ ...valid, email: "admin.com" }).email).toBeDefined();
    });

    it("rejects missing domain", () => {
      expect(validate({ ...valid, email: "admin@" }).email).toBeDefined();
    });

    it("accepts valid email", () => {
      expect(
        validate({ ...valid, email: "user@domain.co" }).email,
      ).toBeUndefined();
    });
  });

  describe("password", () => {
    it("rejects shorter than 5 chars", () => {
      const fields = { ...valid, password: "1234", confirmPassword: "1234" };
      expect(validate(fields).password).toBeDefined();
    });

    it("accepts 5-char password", () => {
      const fields = { ...valid, password: "12345", confirmPassword: "12345" };
      expect(validate(fields).password).toBeUndefined();
    });

    it("accepts 32-char password", () => {
      const pw = "a".repeat(32);
      const fields = { ...valid, password: pw, confirmPassword: pw };
      expect(validate(fields).password).toBeUndefined();
    });

    it("rejects longer than 32 chars", () => {
      const pw = "a".repeat(33);
      const fields = { ...valid, password: pw, confirmPassword: pw };
      expect(validate(fields).password).toBeDefined();
    });
  });

  describe("confirmPassword", () => {
    it("rejects when passwords do not match", () => {
      expect(
        validate({ ...valid, confirmPassword: "different" }).confirmPassword,
      ).toBeDefined();
    });

    it("accepts when passwords match", () => {
      expect(validate(valid).confirmPassword).toBeUndefined();
    });
  });
});
