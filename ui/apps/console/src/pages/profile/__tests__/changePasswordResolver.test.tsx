import { describe, it, expect } from "vitest";
import { changePasswordResolver, type ChangePasswordFormValues } from "../changePasswordResolver";

function resolve(values: Partial<ChangePasswordFormValues>) {
  return changePasswordResolver({
    current: "",
    newPw: "",
    confirmPw: "",
    ...values,
  });
}

describe("changePasswordResolver", () => {
  it("produces no errors when all fields are empty", () => {
    expect(resolve({})).toEqual({});
  });

  it("produces no errors for valid values", () => {
    const result = resolve({ current: "oldpass1", newPw: "newpass123", confirmPw: "newpass123" });
    expect(result).toEqual({});
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
      const result = resolve({ newPw: "newpass123", confirmPw: "different123" });
      expect(result.confirmPw).toBe("Passwords do not match");
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
