import { describe, it, expect } from "vitest";
import { samlSchema, type SamlFormValues } from "../samlSchema";

const validCert = "-----BEGIN CERTIFICATE-----\nMIIBIjANBgkq\n-----END CERTIFICATE-----";
const validUrl = "https://idp.example.com/sso";

/** First validation message per field, mirroring the RHF resolver shape. */
function resolve(
  overrides: Partial<SamlFormValues>,
): Partial<Record<keyof SamlFormValues, string>> {
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

  const result = samlSchema.safeParse({ ...defaults, ...overrides });
  if (result.success) return {};

  const errors: Partial<Record<keyof SamlFormValues, string>> = {};
  for (const issue of result.error.issues) {
    const key = issue.path[0] as keyof SamlFormValues;
    if (key && errors[key] === undefined) errors[key] = issue.message;
  }
  return errors;
}

describe("samlSchema", () => {
  describe("metadata mode (useMetadataUrl: true)", () => {
    it("passes when metadataUrl is a valid URL", () => {
      expect(resolve({ useMetadataUrl: true, metadataUrl: validUrl }).metadataUrl).toBeUndefined();
    });

    it("emits a format error for an invalid metadataUrl when present", () => {
      expect(resolve({ useMetadataUrl: true, metadataUrl: "not-a-url" }).metadataUrl).toBeDefined();
    });

    it("emits an error for an empty metadataUrl", () => {
      expect(resolve({ useMetadataUrl: true, metadataUrl: "" }).metadataUrl).toBeDefined();
    });
  });

  describe("manual mode (useMetadataUrl: false)", () => {
    it("passes when all required fields are valid with a postUrl", () => {
      expect(
        resolve({
          postUrl: validUrl,
          entityId: "https://idp.example.com/entity",
          certificate: validCert,
        }),
      ).toEqual({});
    });

    it("emits an error for an invalid postUrl when present", () => {
      expect(
        resolve({
          postUrl: "bad-url",
          entityId: "https://idp.example.com/entity",
          certificate: validCert,
        }).postUrl,
      ).toBeDefined();
    });

    it("emits an error when entityId is missing", () => {
      expect(
        resolve({
          postUrl: validUrl,
          entityId: "",
          certificate: validCert,
        }).entityId,
      ).toBeDefined();
    });

    it("emits an error for an invalid certificate", () => {
      expect(
        resolve({
          postUrl: validUrl,
          entityId: "https://idp.example.com/entity",
          certificate: "not-a-valid-cert",
        }).certificate,
      ).toBeDefined();
    });

    it("passes when optional URL fields are empty", () => {
      expect(
        resolve({
          postUrl: validUrl,
          redirectUrl: "",
          entityId: "https://idp.example.com/entity",
          certificate: validCert,
        }).redirectUrl,
      ).toBeUndefined();
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
