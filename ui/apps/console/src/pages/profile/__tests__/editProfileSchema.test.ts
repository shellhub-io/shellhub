import { describe, it, expect } from "vitest";
import {
  editProfileSchema,
  type EditProfileFormValues,
} from "../editProfileSchema";

const ALL_CHANGED = { name: "\0", username: "\0", email: "\0" };

/** First validation message per field. */
function resolve(
  values: Partial<EditProfileFormValues>,
  current = ALL_CHANGED,
): Partial<Record<keyof EditProfileFormValues, string>> {
  const result = editProfileSchema(current).safeParse({
    name: "",
    username: "",
    email: "",
    recoveryEmail: "",
    ...values,
  });
  if (result.success) return {};

  const errors: Partial<Record<keyof EditProfileFormValues, string>> = {};
  for (const issue of result.error.issues) {
    const key = issue.path[0] as keyof EditProfileFormValues;
    if (key && errors[key] === undefined) errors[key] = issue.message;
  }
  return errors;
}

describe("editProfileSchema", () => {
  describe("valid values", () => {
    it("produces no errors for a fully valid form", () => {
      expect(
        resolve({
          name: "Alice",
          username: "alice",
          email: "alice@example.com",
          recoveryEmail: "backup@example.com",
        }),
      ).toEqual({});
    });

    it("skips validation for unchanged fields so legacy-invalid values don't block save", () => {
      const legacy = {
        name: "Alice",
        username: "UPPERCASE",
        email: "alice@example.com",
      };
      expect(
        resolve(
          {
            name: "Alice",
            username: "UPPERCASE",
            email: "alice@example.com",
            recoveryEmail: "",
          },
          legacy,
        ),
      ).toEqual({});
    });

    it("produces no errors when recoveryEmail is empty", () => {
      expect(
        resolve({
          name: "Alice",
          username: "alice",
          email: "alice@example.com",
          recoveryEmail: "",
        }),
      ).toEqual({});
    });
  });

  describe("name", () => {
    it("emits 'Name is required' when name is empty", () => {
      expect(resolve({ name: "" }).name).toBe("Name is required");
    });

    it("emits error when name exceeds 64 characters", () => {
      expect(resolve({ name: "a".repeat(65) }).name).toBe(
        "Name must be at most 64 characters",
      );
    });

    it("emits no error for a valid name", () => {
      expect(resolve({ name: "Bob" }).name).toBeUndefined();
    });
  });

  describe("username", () => {
    it("emits 'Username is required' when username is empty", () => {
      expect(resolve({ username: "" }).username).toBe("Username is required");
    });

    it("emits error when username exceeds 32 characters", () => {
      expect(resolve({ username: "a".repeat(33) }).username).toBe(
        "Username must be at most 32 characters",
      );
    });

    it("emits error when username contains uppercase letters", () => {
      expect(resolve({ username: "Alice" }).username).toBe(
        "Username must be lowercase",
      );
    });

    it("emits error when username contains spaces", () => {
      expect(resolve({ username: "alice bob" }).username).toBe(
        "Username cannot contain spaces",
      );
    });

    it("emits error when username contains invalid characters", () => {
      expect(resolve({ username: "alice!" }).username).toBe(
        "Only lowercase letters, numbers, dots, underscores, @ and hyphens are allowed",
      );
    });

    it("emits no error for a valid username", () => {
      expect(resolve({ username: "alice.bob_123" }).username).toBeUndefined();
    });
  });

  describe("email", () => {
    it("emits 'Email is required' when email is empty", () => {
      expect(resolve({ email: "" }).email).toBe("Email is required");
    });

    it("emits error for an invalid email format", () => {
      expect(resolve({ email: "not-an-email" }).email).toBe(
        "Invalid email format",
      );
    });

    it("emits no error for a valid email", () => {
      expect(resolve({ email: "user@example.com" }).email).toBeUndefined();
    });
  });

  describe("recoveryEmail", () => {
    it("emits no error when recoveryEmail is empty", () => {
      expect(resolve({ recoveryEmail: "" }).recoveryEmail).toBeUndefined();
    });

    it("emits error for an invalid recovery email format", () => {
      expect(resolve({ recoveryEmail: "bad-email" }).recoveryEmail).toBe(
        "Invalid email format",
      );
    });

    it("emits 'Must be different from your email' when recoveryEmail matches email", () => {
      expect(
        resolve({
          email: "user@example.com",
          recoveryEmail: "user@example.com",
        }).recoveryEmail,
      ).toBe("Must be different from your email");
    });

    it("emits 'Must be different from your email' case-insensitively", () => {
      expect(
        resolve({
          email: "user@example.com",
          recoveryEmail: "USER@EXAMPLE.COM",
        }).recoveryEmail,
      ).toBe("Must be different from your email");
    });

    it("emits no error for a valid different recovery email", () => {
      expect(
        resolve({
          email: "user@example.com",
          recoveryEmail: "backup@example.com",
        }).recoveryEmail,
      ).toBeUndefined();
    });
  });
});
