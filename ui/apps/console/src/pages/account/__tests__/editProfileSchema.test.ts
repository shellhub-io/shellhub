import { describe, it, expect } from "vitest";
import {
  editProfileSchema,
  type CurrentProfileValues,
  type EditProfileFormValues,
} from "../editProfileSchema";

const ALL_CHANGED = { name: "\0", email: "\0", recoveryEmail: "\0" };
const NO_RECOVERY_EMAIL = { ...ALL_CHANGED, recoveryEmail: "" };

function resolve(
  values: Partial<EditProfileFormValues>,
  current: CurrentProfileValues = ALL_CHANGED,
): Partial<Record<keyof EditProfileFormValues, string>> {
  const result = editProfileSchema(current).safeParse({
    name: "",
    email: "",
    recoveryEmail: "",
    currentPassword: "secret",
    mfaCode: "",
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
          email: "alice@example.com",
          recoveryEmail: "backup@example.com",
        }),
      ).toEqual({});
    });

    it("skips validation for unchanged fields so legacy-invalid values don't block save", () => {
      const legacy = {
        name: "a".repeat(65),
        email: "alice@example.com",
        recoveryEmail: "",
      };
      expect(
        resolve(
          {
            name: "a".repeat(65),
            email: "alice@example.com",
            recoveryEmail: "",
          },
          legacy,
        ),
      ).toEqual({});
    });

    it("produces no errors when recoveryEmail is empty and none is stored", () => {
      expect(
        resolve(
          {
            name: "Alice",
            email: "alice@example.com",
            recoveryEmail: "",
          },
          NO_RECOVERY_EMAIL,
        ),
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

  describe("email", () => {
    it("emits 'Email is required' when email is empty", () => {
      expect(resolve({ email: "" }).email).toBe("Email is required");
    });

    it("emits error for an invalid email format", () => {
      expect(resolve({ email: "not-an-email" }).email).toBe(
        "Enter a valid email address",
      );
    });

    it("emits no error for a valid email", () => {
      expect(resolve({ email: "user@example.com" }).email).toBeUndefined();
    });
  });

  describe("padding", () => {
    const blankName = {
      name: "   ",
      email: "alice@example.com",
      recoveryEmail: "",
    };

    it("rejects a blank stored name once it is edited, even by padding alone", () => {
      expect(resolve({ ...blankName, name: "    " }, blankName).name).toBe(
        "Name is required",
      );
    });
  });

  describe("recoveryEmail", () => {
    it("emits no error when recoveryEmail is empty and none is stored", () => {
      expect(
        resolve({ recoveryEmail: "" }, NO_RECOVERY_EMAIL).recoveryEmail,
      ).toBeUndefined();
    });

    it.each([
      ["empty", ""],
      ["blank", "   "],
    ])("emits 'Recovery email can't be removed' when a stored one is %s", (_, recoveryEmail) => {
      expect(resolve({ recoveryEmail }).recoveryEmail).toBe(
        "Recovery email can't be removed",
      );
    });

    it("emits error for an invalid recovery email format", () => {
      expect(resolve({ recoveryEmail: "bad-email" }).recoveryEmail).toBe(
        "Enter a valid email address",
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

  describe("currentPassword", () => {
    const stored = {
      name: "Alice",
      email: "alice@example.com",
      recoveryEmail: "backup@example.com",
    };
    const unchanged = { ...stored, currentPassword: "" };

    it("is not needed when neither email changes", () => {
      expect(
        resolve({ ...unchanged, name: "Bob" }, stored).currentPassword,
      ).toBeUndefined();
    });

    it.each([
      ["email", { email: "alice@new.example.com" }],
      ["recovery email", { recoveryEmail: "other@example.com" }],
    ])("is required to change the %s", (_, change) => {
      expect(resolve({ ...unchanged, ...change }, stored).currentPassword).toBe(
        "Enter your current password to change an email",
      );
    });

    it.each([
      ["email is cleared", { email: "" }],
      ["recovery email is cleared", { recoveryEmail: "" }],
      ["email is invalid", { email: "not-an-email" }],
    ])("is not needed while the %s", (_, change) => {
      expect(
        resolve({ ...unchanged, ...change }, stored).currentPassword,
      ).toBeUndefined();
    });

    it("is satisfied by any entry, which the server checks", () => {
      expect(
        resolve(
          {
            ...unchanged,
            email: "alice@new.example.com",
            currentPassword: "x",
          },
          stored,
        ).currentPassword,
      ).toBeUndefined();
    });
  });

  describe("mfaCode", () => {
    const stored = {
      name: "Alice",
      email: "alice@example.com",
      recoveryEmail: "backup@example.com",
      mfaEnabled: true,
    };

    it("is not needed when neither email changes", () => {
      expect(
        resolve({ ...stored, name: "Bob" }, stored).mfaCode,
      ).toBeUndefined();
    });

    it("is not needed on an account without MFA", () => {
      expect(
        resolve(
          { ...stored, email: "alice@new.example.com" },
          { ...stored, mfaEnabled: false },
        ).mfaCode,
      ).toBeUndefined();
    });

    it.each([
      ["email", { email: "alice@new.example.com" }, ""],
      ["recovery email", { recoveryEmail: "other@example.com" }, "   "],
    ])("needs a code to change the %s (got %j)", (_, change, mfaCode) => {
      expect(resolve({ ...stored, ...change, mfaCode }, stored).mfaCode).toBe(
        "Enter a code from your authenticator app or a recovery code",
      );
    });

    it.each(["123456", "ABCDEFGHIJKLMNOP"])(
      "is satisfied by %j, which the server checks",
      (mfaCode) => {
        expect(
          resolve({ ...stored, email: "alice@new.example.com", mfaCode }, stored)
            .mfaCode,
        ).toBeUndefined();
      },
    );
  });
});
