import { describe, it, expect } from "vitest";
import { editProfileResolver, type EditProfileFormValues } from "../editProfileResolver";

function resolve(values: Partial<EditProfileFormValues>) {
  return editProfileResolver({
    name: "",
    username: "",
    email: "",
    recoveryEmail: "",
    ...values,
  });
}

describe("editProfileResolver", () => {
  describe("valid values", () => {
    it("produces no errors for a fully valid form", () => {
      const result = resolve({
        name: "Alice",
        username: "alice",
        email: "alice@example.com",
        recoveryEmail: "backup@example.com",
      });
      expect(result).toEqual({});
    });

    it("produces no errors when recoveryEmail is empty", () => {
      const result = resolve({
        name: "Alice",
        username: "alice",
        email: "alice@example.com",
        recoveryEmail: "",
      });
      expect(result).toEqual({});
    });
  });

  describe("name", () => {
    it("emits no error when name is empty (not present-but-invalid)", () => {
      const result = resolve({ name: "" });
      expect(result.name).toBeUndefined();
    });

    it("emits error when name exceeds 64 characters", () => {
      const result = resolve({ name: "a".repeat(65) });
      expect(result.name).toBe("Name must be at most 64 characters");
    });

    it("emits no error for a valid name", () => {
      const result = resolve({ name: "Bob" });
      expect(result.name).toBeUndefined();
    });
  });

  describe("username", () => {
    it("emits no error when username is empty", () => {
      const result = resolve({ username: "" });
      expect(result.username).toBeUndefined();
    });

    it("emits error when username exceeds 32 characters", () => {
      const result = resolve({ username: "a".repeat(33) });
      expect(result.username).toBe("Username must be at most 32 characters");
    });

    it("emits error when username contains uppercase letters", () => {
      const result = resolve({ username: "Alice" });
      expect(result.username).toBe("Username must be lowercase");
    });

    it("emits error when username contains spaces", () => {
      const result = resolve({ username: "alice bob" });
      expect(result.username).toBe("Username cannot contain spaces");
    });

    it("emits error when username contains invalid characters", () => {
      const result = resolve({ username: "alice!" });
      expect(result.username).toBe(
        "Only lowercase letters, numbers, dots, underscores, @ and hyphens are allowed",
      );
    });

    it("emits no error for a valid username", () => {
      const result = resolve({ username: "alice.bob_123" });
      expect(result.username).toBeUndefined();
    });
  });

  describe("email", () => {
    it("emits no error when email is empty", () => {
      const result = resolve({ email: "" });
      expect(result.email).toBeUndefined();
    });

    it("emits error for an invalid email format", () => {
      const result = resolve({ email: "not-an-email" });
      expect(result.email).toBe("Invalid email format");
    });

    it("emits no error for a valid email", () => {
      const result = resolve({ email: "user@example.com" });
      expect(result.email).toBeUndefined();
    });
  });

  describe("recoveryEmail", () => {
    it("emits no error when recoveryEmail is empty", () => {
      const result = resolve({ recoveryEmail: "" });
      expect(result.recoveryEmail).toBeUndefined();
    });

    it("emits error for an invalid recovery email format", () => {
      const result = resolve({ recoveryEmail: "bad-email" });
      expect(result.recoveryEmail).toBe("Invalid email format");
    });

    it("emits 'Must be different from your email' when recoveryEmail matches email", () => {
      const result = resolve({
        email: "user@example.com",
        recoveryEmail: "user@example.com",
      });
      expect(result.recoveryEmail).toBe("Must be different from your email");
    });

    it("emits 'Must be different from your email' case-insensitively", () => {
      const result = resolve({
        email: "user@example.com",
        recoveryEmail: "USER@EXAMPLE.COM",
      });
      expect(result.recoveryEmail).toBe("Must be different from your email");
    });

    it("emits no error for a valid different recovery email", () => {
      const result = resolve({
        email: "user@example.com",
        recoveryEmail: "backup@example.com",
      });
      expect(result.recoveryEmail).toBeUndefined();
    });
  });
});
