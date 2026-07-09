import { describe, it, expect } from "vitest";
import {
  changePasswordSchema,
  type ChangePasswordFormValues,
} from "../changePasswordSchema";

/** First validation message per field. */
function resolve(
  values: Partial<ChangePasswordFormValues>,
): Partial<Record<keyof ChangePasswordFormValues, string>> {
  const result = changePasswordSchema.safeParse({
    current: "",
    newPw: "",
    confirmPw: "",
    ...values,
  });
  if (result.success) return {};

  const errors: Partial<Record<keyof ChangePasswordFormValues, string>> = {};
  for (const issue of result.error.issues) {
    const key = issue.path[0] as keyof ChangePasswordFormValues;
    if (key && errors[key] === undefined) errors[key] = issue.message;
  }
  return errors;
}

describe("changePasswordSchema", () => {
  it("produces no errors when all fields are empty", () => {
    expect(resolve({})).toEqual({});
  });

  it("produces no errors for valid values", () => {
    expect(
      resolve({ current: "oldpass1", newPw: "newpass123", confirmPw: "newpass123" }),
    ).toEqual({});
  });

  describe("newPw", () => {
    it("emits a validatePassword error when newPw is non-empty and invalid", () => {
      const result = resolve({ newPw: "x", confirmPw: "x" });
      expect(result.newPw).toBeDefined();
      expect(typeof result.newPw).toBe("string");
    });

    it("emits no error when newPw is empty", () => {
      expect(resolve({ newPw: "" }).newPw).toBeUndefined();
    });
  });

  describe("confirmPw", () => {
    it("emits 'Passwords do not match' when confirmPw is non-empty and mismatches newPw", () => {
      expect(resolve({ newPw: "newpass123", confirmPw: "different123" }).confirmPw).toBe(
        "Passwords do not match",
      );
    });

    it("emits no error when confirmPw is empty", () => {
      expect(resolve({ newPw: "newpass123", confirmPw: "" }).confirmPw).toBeUndefined();
    });
  });

  describe("current", () => {
    it("emits no error for empty current password", () => {
      expect(resolve({ current: "" }).current).toBeUndefined();
    });
  });
});
