import { z } from "zod";
import type { GetAuthenticationSettingsResponse } from "@/client";

/**
 * The stored SAML configuration, taken from the generated response type so the form and the API
 * cannot drift.
 */
export type SamlSettings = NonNullable<GetAuthenticationSettingsResponse>["saml"];

function isValidUrl(s: string): boolean {
  try {
    new URL(s);
    return true;
  } catch {
    return false;
  }
}

function isCertValid(s: string): boolean {
  return (
    s.includes("-----BEGIN CERTIFICATE-----") &&
    s.includes("-----END CERTIFICATE-----")
  );
}

/**
 * Normalizes a pasted certificate: providers hand out PEM with varying line breaks and stray
 * whitespace, and the server wants one consistent form. Input without the PEM markers is
 * returned untouched, so a partial paste is reported by validation rather than mangled here.
 */
export function normalizeCert(raw: string): string {
  const begin = "-----BEGIN CERTIFICATE-----";
  const end = "-----END CERTIFICATE-----";
  if (!raw.includes(begin) || !raw.includes(end)) return raw;
  const body = raw
    .slice(raw.indexOf(begin) + begin.length, raw.lastIndexOf(end))
    .replace(/\s+/g, "");
  return `${begin}\n${body}\n${end}`;
}

/**
 * Validates the SAML form. The metadata URL and the manual fields are alternatives, so exactly
 * one branch is required and the other is left alone.
 */
export const samlSchema = z
  .object({
    useMetadataUrl: z.boolean(),
    metadataUrl: z.string(),
    postUrl: z.string(),
    redirectUrl: z.string(),
    entityId: z.string(),
    certificate: z.string(),
    emailMapping: z.string(),
    nameMapping: z.string(),
    signRequests: z.boolean(),
  })
  .superRefine((values, ctx) => {
    if (values.useMetadataUrl) {
      if (!values.metadataUrl || !isValidUrl(values.metadataUrl)) {
        ctx.addIssue({ code: "custom", path: ["metadataUrl"], message: "Must be a valid URL" });
      }
      return;
    }

    if (!values.postUrl && !values.redirectUrl) {
      ctx.addIssue({
        code: "custom",
        path: ["postUrl"],
        message: "At least one Sign-On URL (POST or Redirect) is required",
      });
    } else {
      if (values.postUrl && !isValidUrl(values.postUrl)) {
        ctx.addIssue({ code: "custom", path: ["postUrl"], message: "Must be a valid URL" });
      }
      if (values.redirectUrl && !isValidUrl(values.redirectUrl)) {
        ctx.addIssue({ code: "custom", path: ["redirectUrl"], message: "Must be a valid URL" });
      }
    }

    if (!values.entityId.trim()) {
      ctx.addIssue({ code: "custom", path: ["entityId"], message: "Entity ID is required" });
    }

    if (!values.certificate.trim()) {
      ctx.addIssue({ code: "custom", path: ["certificate"], message: "Certificate is required" });
    } else if (!isCertValid(values.certificate)) {
      ctx.addIssue({
        code: "custom",
        path: ["certificate"],
        message: "Must include BEGIN CERTIFICATE and END CERTIFICATE markers",
      });
    }
  });

/**
 * The SAML form's values, derived from the schema.
 */
export type SamlFormValues = z.infer<typeof samlSchema>;

/**
 * Fills the SAML form from the stored configuration, or with blanks when there is none.
 */
export function buildSamlDefaults(
  existingConfig: SamlSettings | null | undefined,
): SamlFormValues {
  if (!existingConfig?.enabled) {
    return {
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
  }

  return {
    useMetadataUrl: false,
    metadataUrl: "",
    postUrl: existingConfig.idp?.binding?.post ?? "",
    redirectUrl: existingConfig.idp?.binding?.redirect ?? "",
    entityId: existingConfig.idp?.entity_id ?? "",
    certificate: existingConfig.idp?.certificates?.[0] ?? "",
    emailMapping: existingConfig.idp?.mappings?.email ?? "",
    nameMapping: existingConfig.idp?.mappings?.name ?? "",
    signRequests: existingConfig.sp?.sign_requests ?? false,
  };
}

/**
 * Turns validated SAML form values into the request body, sending only the branch that was
 * filled in — the metadata URL or the manual fields, never both.
 */
export function buildSamlBody(values: SamlFormValues) {
  if (values.useMetadataUrl) {
    return {
      enable: true,
      idp: { metadata_url: values.metadataUrl },
      sp: { sign_requests: values.signRequests },
    };
  }
  return {
    enable: true,
    idp: {
      entity_id: values.entityId,
      binding: {
        ...(values.postUrl ? { post: values.postUrl } : {}),
        ...(values.redirectUrl ? { redirect: values.redirectUrl } : {}),
      },
      certificate: normalizeCert(values.certificate),
      ...(values.emailMapping || values.nameMapping
        ? {
            mappings: {
              ...(values.emailMapping ? { email: values.emailMapping } : {}),
              ...(values.nameMapping ? { name: values.nameMapping } : {}),
            },
          }
        : {}),
    },
    sp: { sign_requests: values.signRequests },
  };
}
