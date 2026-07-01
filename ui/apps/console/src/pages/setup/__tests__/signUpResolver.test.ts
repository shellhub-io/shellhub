import { describe, it, expect } from "vitest";
import { signUpResolver } from "../signUpResolver";
import type { SignUpFormValues } from "../signUpResolver";

const validInput: SignUpFormValues = {
  name: "Admin User",
  username: "admin",
  email: "admin@example.com",
  password: "secret123",
  confirmPassword: "secret123",
  acceptPrivacyPolicy: true,
};

const emptyOptions = {
  criteriaMode: undefined,
  fields: {},
  names: undefined,
  shouldUseNativeValidation: undefined,
} as Parameters<typeof signUpResolver>[2];

describe("signUpResolver", () => {
  it("returns values with no errors for valid input", async () => {
    const result = await signUpResolver(validInput, undefined, emptyOptions);
    expect(result.values).toEqual(validInput);
    expect(result.errors).toEqual({});
  });

  describe("validate() field error mapping", () => {
    it("maps name error to RHF error shape", async () => {
      const input: SignUpFormValues = { ...validInput, name: "" };
      const result = await signUpResolver(input, undefined, emptyOptions);
      expect(result.errors).toHaveProperty("name");
      expect((result.errors as Record<string, { type: string; message: string }>).name.type).toBe("validate");
      expect(typeof (result.errors as Record<string, { type: string; message: string }>).name.message).toBe("string");
      expect(result.values).toEqual({});
    });

    it("maps username error to RHF error shape", async () => {
      const input: SignUpFormValues = { ...validInput, username: "ab" };
      const result = await signUpResolver(input, undefined, emptyOptions);
      expect(result.errors).toHaveProperty("username");
      expect((result.errors as Record<string, { type: string; message: string }>).username.type).toBe("validate");
      expect(typeof (result.errors as Record<string, { type: string; message: string }>).username.message).toBe("string");
      expect(result.values).toEqual({});
    });

    it("maps email error to RHF error shape", async () => {
      const input: SignUpFormValues = { ...validInput, email: "not-an-email" };
      const result = await signUpResolver(input, undefined, emptyOptions);
      expect(result.errors).toHaveProperty("email");
      expect((result.errors as Record<string, { type: string; message: string }>).email.type).toBe("validate");
      expect(typeof (result.errors as Record<string, { type: string; message: string }>).email.message).toBe("string");
      expect(result.values).toEqual({});
    });

    it("maps password error to RHF error shape", async () => {
      const input: SignUpFormValues = { ...validInput, password: "123", confirmPassword: "123" };
      const result = await signUpResolver(input, undefined, emptyOptions);
      expect(result.errors).toHaveProperty("password");
      expect((result.errors as Record<string, { type: string; message: string }>).password.type).toBe("validate");
      expect(typeof (result.errors as Record<string, { type: string; message: string }>).password.message).toBe("string");
      expect(result.values).toEqual({});
    });

    it("maps confirmPassword error to RHF error shape", async () => {
      const input: SignUpFormValues = { ...validInput, confirmPassword: "different" };
      const result = await signUpResolver(input, undefined, emptyOptions);
      expect(result.errors).toHaveProperty("confirmPassword");
      expect((result.errors as Record<string, { type: string; message: string }>).confirmPassword.type).toBe("validate");
      expect(typeof (result.errors as Record<string, { type: string; message: string }>).confirmPassword.message).toBe("string");
      expect(result.values).toEqual({});
    });
  });

  describe("acceptPrivacyPolicy", () => {
    it("returns error on acceptPrivacyPolicy when false", async () => {
      const input: SignUpFormValues = { ...validInput, acceptPrivacyPolicy: false };
      const result = await signUpResolver(input, undefined, emptyOptions);
      expect(result.errors).toHaveProperty("acceptPrivacyPolicy");
      expect((result.errors as Record<string, { type: string; message: string }>).acceptPrivacyPolicy.type).toBe("required");
      expect(result.values).toEqual({});
    });

    it("clears acceptPrivacyPolicy error when true", async () => {
      const result = await signUpResolver(validInput, undefined, emptyOptions);
      expect(result.errors).not.toHaveProperty("acceptPrivacyPolicy");
    });
  });
});
