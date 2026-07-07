import { describe, it, expect } from "vitest";
import { samlResolver, type SamlFormValues } from "../samlResolver";

const validCert = "-----BEGIN CERTIFICATE-----\nMIIBIjANBgkq\n-----END CERTIFICATE-----";
const validUrl = "https://idp.example.com/sso";

function resolve(overrides: Partial<SamlFormValues>) {
  const defaults: SamlFormValues = {
    useMetadataUrl: false,
    metadataUrl: "",
    postUrl: "",
    redirectUrl: "",
    entityId: "",
    certificate: "",
    emailMapping: "",
    nameMapping: "",
    signRequests: false,
  };
  return samlResolver({ ...defaults, ...overrides });
}

describe("samlResolver", () => {
  describe("metadata mode (useMetadataUrl: true)", () => {
    it("passes when metadataUrl is a valid URL", () => {
      const result = resolve({ useMetadataUrl: true, metadataUrl: validUrl });
      expect(result.metadataUrl).toBeUndefined();
    });

    it("emits a format error for an invalid metadataUrl when present", () => {
      const result = resolve({ useMetadataUrl: true, metadataUrl: "not-a-url" });
      expect(result.metadataUrl).toBeDefined();
    });

    it("emits an error for an empty metadataUrl", () => {
      const result = resolve({ useMetadataUrl: true, metadataUrl: "" });
      expect(result.metadataUrl).toBeDefined();
    });
  });

  describe("manual mode (useMetadataUrl: false)", () => {
    it("passes when all required fields are valid with a postUrl", () => {
      const result = resolve({
        postUrl: validUrl,
        entityId: "https://idp.example.com/entity",
        certificate: validCert,
      });
      expect(result).toEqual({});
    });

    it("emits an error for an invalid postUrl when present", () => {
      const result = resolve({
        postUrl: "bad-url",
        entityId: "https://idp.example.com/entity",
        certificate: validCert,
      });
      expect(result.postUrl).toBeDefined();
    });

    it("emits an error when entityId is missing", () => {
      const result = resolve({
        postUrl: validUrl,
        entityId: "",
        certificate: validCert,
      });
      expect(result.entityId).toBeDefined();
    });

    it("emits an error for an invalid certificate", () => {
      const result = resolve({
        postUrl: validUrl,
        entityId: "https://idp.example.com/entity",
        certificate: "not-a-valid-cert",
      });
      expect(result.certificate).toBeDefined();
    });

    it("passes when optional URL fields are empty", () => {
      const result = resolve({
        postUrl: validUrl,
        redirectUrl: "",
        entityId: "https://idp.example.com/entity",
        certificate: validCert,
      });
      expect(result.redirectUrl).toBeUndefined();
    });

    it("produces no error for blank non-required fields (present-only emission)", () => {
      const result = resolve({
        postUrl: validUrl,
        entityId: "https://idp.example.com/entity",
        certificate: validCert,
        emailMapping: "",
        nameMapping: "",
      });
      expect(result.emailMapping).toBeUndefined();
      expect(result.nameMapping).toBeUndefined();
    });
  });
});
