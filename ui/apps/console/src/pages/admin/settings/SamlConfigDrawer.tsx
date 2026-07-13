import { useState } from "react";
import {
  KeyIcon,
  ChevronDownIcon,
  ExclamationCircleIcon,
} from "@heroicons/react/24/outline";
import { configureSamlAuthentication } from "@/client";
import type { GetAuthenticationSettingsResponse } from "@/client";
import { cn } from "@shellhub/design-system/cn";
import Drawer from "@/components/common/Drawer";
import InputField from "@/components/common/fields/InputField";
import CheckboxField from "@/components/common/fields/CheckboxField";
import { INPUT } from "@/utils/styles";
import FieldLabel from "@/components/common/fields/FieldLabel";
import { Button } from "@shellhub/design-system/primitives";

type SamlSettings = NonNullable<GetAuthenticationSettingsResponse>["saml"];

interface Props {
  open: boolean;
  onClose: () => void;
  onSaved: () => void;
  existingConfig: SamlSettings | null | undefined;
}

interface FormState {
  useMetadataUrl: boolean;
  metadataUrl: string;
  postUrl: string;
  redirectUrl: string;
  entityId: string;
  certificate: string;
  emailMapping: string;
  nameMapping: string;
  signRequests: boolean;
  showAdvanced: boolean;
  saving: boolean;
  formError: string | null;
}

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

function normalizeCert(raw: string): string {
  const begin = "-----BEGIN CERTIFICATE-----";
  const end = "-----END CERTIFICATE-----";
  if (!raw.includes(begin) || !raw.includes(end)) return raw;
  const body = raw
    .slice(raw.indexOf(begin) + begin.length, raw.lastIndexOf(end))
    .replace(/\s+/g, "");
  return `${begin}\n${body}\n${end}`;
}

function buildInitialState(
  existingConfig: SamlSettings | null | undefined,
): FormState {
  const base: FormState = {
    useMetadataUrl: false,
    metadataUrl: "",
    postUrl: "",
    redirectUrl: "",
    entityId: "",
    certificate: "",
    emailMapping: "",
    nameMapping: "",
    signRequests: false,
    showAdvanced: false,
    saving: false,
    formError: null,
  };

  if (!existingConfig?.enabled) return base;

  return {
    ...base,
    postUrl: existingConfig.idp?.binding?.post ?? "",
    redirectUrl: existingConfig.idp?.binding?.redirect ?? "",
    entityId: existingConfig.idp?.entity_id ?? "",
    certificate: existingConfig.idp?.certificates?.[0] ?? "",
    signRequests: existingConfig.sp?.sign_requests ?? false,
    emailMapping: existingConfig.idp?.mappings?.email ?? "",
    nameMapping: existingConfig.idp?.mappings?.name ?? "",
    showAdvanced: !!(
      existingConfig.idp?.mappings?.email || existingConfig.idp?.mappings?.name
    ),
  };
}

export default function SamlConfigDrawer({
  open,
  onClose,
  onSaved,
  existingConfig,
}: Props) {
  const [form, setForm] = useState<FormState>(() =>
    buildInitialState(existingConfig),
  );
  const [prevOpen, setPrevOpen] = useState(open);

  // Populate / reset when drawer (re-)opens.
  // Done during render (not in an effect) so React can discard this render
  // and immediately re-render with the fresh state — no cascading renders.
  if (open !== prevOpen) {
    setPrevOpen(open);
    if (open) {
      setForm(buildInitialState(existingConfig));
    }
  }

  const {
    useMetadataUrl,
    metadataUrl,
    postUrl,
    redirectUrl,
    entityId,
    certificate,
    emailMapping,
    nameMapping,
    signRequests,
    showAdvanced,
    saving,
    formError,
  } = form;

  const patch = (fields: Partial<FormState>) =>
    setForm((prev) => ({ ...prev, ...fields }));

  // Validation
  const noUrlProvided = !useMetadataUrl && !postUrl && !redirectUrl;
  const postUrlInvalid = !useMetadataUrl && !!postUrl && !isValidUrl(postUrl);
  const redirectUrlInvalid =
    !useMetadataUrl && !!redirectUrl && !isValidUrl(redirectUrl);
  const certInvalid =
    !useMetadataUrl && !!certificate && !isCertValid(certificate);

  const hasErrors = useMetadataUrl
    ? !metadataUrl || !isValidUrl(metadataUrl)
    : noUrlProvided ||
      postUrlInvalid ||
      redirectUrlInvalid ||
      !entityId.trim() ||
      !certificate.trim() ||
      certInvalid;

  const buildBody = () => {
    if (useMetadataUrl) {
      return {
        enable: true,
        idp: { metadata_url: metadataUrl },
        sp: { sign_requests: signRequests },
      };
    }
    return {
      enable: true,
      idp: {
        entity_id: entityId,
        binding: {
          ...(postUrl ? { post: postUrl } : {}),
          ...(redirectUrl ? { redirect: redirectUrl } : {}),
        },
        certificate: normalizeCert(certificate),
        ...(emailMapping || nameMapping
          ? {
              mappings: {
                ...(emailMapping ? { email: emailMapping } : {}),
                ...(nameMapping ? { name: nameMapping } : {}),
              },
            }
          : {}),
      },
      sp: { sign_requests: signRequests },
    };
  };

  const handleSave = async () => {
    patch({ saving: true, formError: null });
    try {
      await configureSamlAuthentication({
        body: buildBody(),
        throwOnError: true,
      });
      onSaved();
      onClose();
    } catch {
      patch({
        formError:
          "Failed to save SAML configuration. Please check your settings and try again.",
      });
    } finally {
      patch({ saving: false });
    }
  };

  const footer = (
    <>
      <Button variant="ghost" onClick={onClose}>
        Cancel
      </Button>
      <Button
        loading={saving}
        disabled={hasErrors}
        onClick={() => void handleSave()}
      >
        Save Configuration
      </Button>
    </>
  );

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Configure Single Sign-On"
      subtitle="Configure SAML authentication for your ShellHub instance"
      icon={<KeyIcon className="w-4 h-4 text-primary" />}
      width="md"
      footer={footer}
    >
      <div className="space-y-5">
        {formError && (
          <div
            role="alert"
            className="flex items-start gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3.5 py-2.5 rounded-md text-xs font-mono"
          >
            <ExclamationCircleIcon
              className="w-3.5 h-3.5 shrink-0 mt-0.5"
              strokeWidth={2}
            />
            {formError}
          </div>
        )}

        {/* Metadata URL toggle */}
        <CheckboxField
          id="saml-use-metadata-url"
          label="Use Metadata URL"
          description="Automatically fetch IdP configuration from a URL"
          checked={useMetadataUrl}
          onChange={(checked) => patch({ useMetadataUrl: checked })}
        />

        {useMetadataUrl ? (
          /* Metadata URL mode */
          <InputField
            id="metadata-url"
            label="IdP Metadata URL"
            type="url"
            value={metadataUrl}
            onChange={(v) => patch({ metadataUrl: v })}
            placeholder="https://idp.example.com/metadata.xml"
            variant="mono"
            error={
              metadataUrl && !isValidUrl(metadataUrl)
                ? "Must be a valid URL"
                : undefined
            }
          />
        ) : (
          /* Manual configuration mode */
          <div className="space-y-4">
            {/* URL warning */}
            {noUrlProvided && (
              <div className="flex items-start gap-2 bg-accent-yellow/8 border border-accent-yellow/20 text-accent-yellow px-3.5 py-2.5 rounded-md text-xs font-mono">
                <ExclamationCircleIcon
                  className="w-3.5 h-3.5 shrink-0 mt-0.5"
                  strokeWidth={2}
                />
                Please provide at least one Sign-On URL (POST or Redirect)
              </div>
            )}

            <InputField
              id="post-url"
              label="SSO POST URL"
              type="url"
              value={postUrl}
              onChange={(v) => patch({ postUrl: v })}
              placeholder="https://idp.example.com/sso/post"
              variant="mono"
              error={postUrlInvalid ? "Must be a valid URL" : undefined}
            />

            <InputField
              id="redirect-url"
              label="SSO Redirect URL"
              type="url"
              value={redirectUrl}
              onChange={(v) => patch({ redirectUrl: v })}
              placeholder="https://idp.example.com/sso/redirect"
              variant="mono"
              error={redirectUrlInvalid ? "Must be a valid URL" : undefined}
            />

            <InputField
              id="entity-id"
              label="Entity ID"
              value={entityId}
              onChange={(v) => patch({ entityId: v })}
              placeholder="https://idp.example.com/entity"
              variant="mono"
              hint="Issuer/Entity ID from your IdP's SAML configuration"
            />

            <div>
              <FieldLabel htmlFor="certificate">X.509 Certificate</FieldLabel>
              <textarea
                id="certificate"
                value={certificate}
                onChange={(e) => patch({ certificate: e.target.value })}
                rows={5}
                placeholder={
                  "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"
                }
                aria-invalid={certInvalid || undefined}
                aria-describedby={certInvalid ? "certificate-error" : undefined}
                className={cn(INPUT, "resize-none font-mono text-2xs leading-relaxed", certInvalid && "border-accent-red/50")}
              />
              {certInvalid && (
                <p
                  id="certificate-error"
                  className="mt-1 text-2xs text-accent-red font-mono"
                >
                  Must include BEGIN CERTIFICATE and END CERTIFICATE markers
                </p>
              )}
            </div>
          </div>
        )}

        {/* Advanced settings (collapsible) */}
        <div className="border border-border rounded-lg overflow-hidden">
          <button
            type="button"
            onClick={() => patch({ showAdvanced: !showAdvanced })}
            className="w-full flex items-center justify-between px-4 py-3 text-sm font-medium text-text-secondary hover:text-text-primary hover:bg-hover-subtle transition-colors"
          >
            <span>Advanced Settings</span>
            <ChevronDownIcon
              className={cn("w-4 h-4 transition-transform duration-200", showAdvanced && "rotate-180")}
              strokeWidth={2}
            />
          </button>

          {showAdvanced && (
            <div className="px-4 pb-4 pt-1 space-y-4 border-t border-border">
              <InputField
                id="email-mapping"
                label="Email Attribute Mapping"
                value={emailMapping}
                onChange={(v) => patch({ emailMapping: v })}
                placeholder="emailAddress"
                variant="mono"
                hint="SAML attribute name that contains the user's email address"
              />

              <InputField
                id="name-mapping"
                label="Name Attribute Mapping"
                value={nameMapping}
                onChange={(v) => patch({ nameMapping: v })}
                placeholder="displayName"
                variant="mono"
                hint="SAML attribute name that contains the user's display name"
              />

              <CheckboxField
                id="saml-sign-requests"
                label="Sign authorization requests"
                description="Allows the IdP to verify that SAML requests originated from ShellHub"
                checked={signRequests}
                onChange={(checked) => patch({ signRequests: checked })}
              />
            </div>
          )}
        </div>
      </div>
    </Drawer>
  );
}
