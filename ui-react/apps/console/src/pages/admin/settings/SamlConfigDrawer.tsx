import { useState } from "react";
import {
  KeyIcon,
  ChevronDownIcon,
  ExclamationCircleIcon,
} from "@heroicons/react/24/outline";
import { configureSamlAuthentication } from "../../../client";
import type { GetAuthenticationSettingsResponse } from "../../../client";
import Drawer from "../../../components/common/Drawer";

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
    s.includes("-----BEGIN CERTIFICATE-----")
    && s.includes("-----END CERTIFICATE-----")
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

function buildInitialState(existingConfig: SamlSettings | null | undefined): FormState {
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
    showAdvanced: !!(existingConfig.idp?.mappings?.email || existingConfig.idp?.mappings?.name),
  };
}

const fieldLabel = "block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-2";
const fieldInput = "w-full px-3 py-2.5 bg-background border border-border rounded-lg text-sm text-text-primary font-mono placeholder:text-text-secondary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all duration-200";
const fieldInputError = "border-accent-red/50 focus:border-accent-red/50 focus:ring-accent-red/20";

export default function SamlConfigDrawer({
  open,
  onClose,
  onSaved,
  existingConfig,
}: Props) {
  const [form, setForm] = useState<FormState>(() => buildInitialState(existingConfig));
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

  const { useMetadataUrl, metadataUrl, postUrl, redirectUrl, entityId,
    certificate, emailMapping, nameMapping, signRequests,
    showAdvanced, saving, formError } = form;

  const patch = (fields: Partial<FormState>) => setForm((prev) => ({ ...prev, ...fields }));

  // Validation
  const noUrlProvided = !useMetadataUrl && !postUrl && !redirectUrl;
  const postUrlInvalid = !useMetadataUrl && !!postUrl && !isValidUrl(postUrl);
  const redirectUrlInvalid = !useMetadataUrl && !!redirectUrl && !isValidUrl(redirectUrl);
  const certInvalid = !useMetadataUrl && !!certificate && !isCertValid(certificate);

  const hasErrors = useMetadataUrl
    ? !metadataUrl || !isValidUrl(metadataUrl)
    : noUrlProvided
      || postUrlInvalid
      || redirectUrlInvalid
      || !entityId.trim()
      || !certificate.trim()
      || certInvalid;

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
        formError: "Failed to save SAML configuration. Please check your settings and try again.",
      });
    } finally {
      patch({ saving: false });
    }
  };

  const footer = (
    <>
      <button
        type="button"
        onClick={onClose}
        className="px-4 py-2 text-sm font-medium text-text-secondary hover:text-text-primary border border-border hover:border-border-light rounded-lg transition-all"
      >
        Cancel
      </button>
      <button
        type="button"
        onClick={() => void handleSave()}
        disabled={hasErrors || saving}
        className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all"
      >
        {saving && (
          <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
        )}
        Save Configuration
      </button>
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
            <ExclamationCircleIcon className="w-3.5 h-3.5 shrink-0 mt-0.5" strokeWidth={2} />
            {formError}
          </div>
        )}

        {/* Metadata URL toggle */}
        <label className="flex items-center gap-3 cursor-pointer select-none">
          <input
            type="checkbox"
            checked={useMetadataUrl}
            onChange={(e) => patch({ useMetadataUrl: e.target.checked })}
            className="w-4 h-4 rounded border-border bg-background text-primary focus:ring-primary/30"
          />
          <span className="text-sm text-text-primary font-medium">Use Metadata URL</span>
          <span className="text-2xs text-text-muted">
            Automatically fetch IdP configuration from a URL
          </span>
        </label>

        {useMetadataUrl ? (
          /* Metadata URL mode */
          <div>
            <label htmlFor="metadata-url" className={fieldLabel}>
              IdP Metadata URL
            </label>
            <input
              id="metadata-url"
              type="url"
              value={metadataUrl}
              onChange={(e) => patch({ metadataUrl: e.target.value })}
              placeholder="https://idp.example.com/metadata.xml"
              className={`${fieldInput} ${metadataUrl && !isValidUrl(metadataUrl) ? fieldInputError : ""}`}
            />
            {metadataUrl && !isValidUrl(metadataUrl) && (
              <p className="mt-1 text-2xs text-accent-red font-mono">Must be a valid URL</p>
            )}
          </div>
        ) : (
          /* Manual configuration mode */
          <div className="space-y-4">
            {/* URL warning */}
            {noUrlProvided && (
              <div className="flex items-start gap-2 bg-accent-yellow/8 border border-accent-yellow/20 text-accent-yellow px-3.5 py-2.5 rounded-md text-xs font-mono">
                <ExclamationCircleIcon className="w-3.5 h-3.5 shrink-0 mt-0.5" strokeWidth={2} />
                Please provide at least one Sign-On URL (POST or Redirect)
              </div>
            )}

            <div>
              <label htmlFor="post-url" className={fieldLabel}>
                SSO POST URL
              </label>
              <input
                id="post-url"
                type="url"
                value={postUrl}
                onChange={(e) => patch({ postUrl: e.target.value })}
                placeholder="https://idp.example.com/sso/post"
                className={`${fieldInput} ${postUrlInvalid ? fieldInputError : ""}`}
              />
              {postUrlInvalid && (
                <p className="mt-1 text-2xs text-accent-red font-mono">Must be a valid URL</p>
              )}
            </div>

            <div>
              <label htmlFor="redirect-url" className={fieldLabel}>
                SSO Redirect URL
              </label>
              <input
                id="redirect-url"
                type="url"
                value={redirectUrl}
                onChange={(e) => patch({ redirectUrl: e.target.value })}
                placeholder="https://idp.example.com/sso/redirect"
                className={`${fieldInput} ${redirectUrlInvalid ? fieldInputError : ""}`}
              />
              {redirectUrlInvalid && (
                <p className="mt-1 text-2xs text-accent-red font-mono">Must be a valid URL</p>
              )}
            </div>

            <div>
              <label htmlFor="entity-id" className={fieldLabel}>
                Entity ID
              </label>
              <input
                id="entity-id"
                type="text"
                value={entityId}
                onChange={(e) => patch({ entityId: e.target.value })}
                placeholder="https://idp.example.com/entity"
                className={fieldInput}
              />
              <p className="mt-1 text-2xs text-text-muted font-mono">
                Issuer/Entity ID from your IdP&apos;s SAML configuration
              </p>
            </div>

            <div>
              <label htmlFor="certificate" className={fieldLabel}>
                X.509 Certificate
              </label>
              <textarea
                id="certificate"
                value={certificate}
                onChange={(e) => patch({ certificate: e.target.value })}
                rows={5}
                placeholder={"-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"}
                className={`${fieldInput} resize-none font-mono text-2xs leading-relaxed ${certInvalid ? fieldInputError : ""}`}
              />
              {certInvalid && (
                <p className="mt-1 text-2xs text-accent-red font-mono">
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
              className={`w-4 h-4 transition-transform duration-200 ${showAdvanced ? "rotate-180" : ""}`}
              strokeWidth={2}
            />
          </button>

          {showAdvanced && (
            <div className="px-4 pb-4 pt-1 space-y-4 border-t border-border">
              <div>
                <label htmlFor="email-mapping" className={fieldLabel}>
                  Email Attribute Mapping
                </label>
                <input
                  id="email-mapping"
                  type="text"
                  value={emailMapping}
                  onChange={(e) => patch({ emailMapping: e.target.value })}
                  placeholder="emailAddress"
                  className={fieldInput}
                />
                <p className="mt-1 text-2xs text-text-muted font-mono">
                  SAML attribute name that contains the user&apos;s email address
                </p>
              </div>

              <div>
                <label htmlFor="name-mapping" className={fieldLabel}>
                  Name Attribute Mapping
                </label>
                <input
                  id="name-mapping"
                  type="text"
                  value={nameMapping}
                  onChange={(e) => patch({ nameMapping: e.target.value })}
                  placeholder="displayName"
                  className={fieldInput}
                />
                <p className="mt-1 text-2xs text-text-muted font-mono">
                  SAML attribute name that contains the user&apos;s display name
                </p>
              </div>

              <label className="flex items-start gap-3 cursor-pointer select-none">
                <input
                  type="checkbox"
                  checked={signRequests}
                  onChange={(e) => patch({ signRequests: e.target.checked })}
                  className="mt-0.5 w-4 h-4 rounded border-border bg-background text-primary focus:ring-primary/30"
                />
                <div>
                  <span className="text-sm text-text-primary font-medium block">
                    Sign authorization requests
                  </span>
                  <span className="text-2xs text-text-muted font-mono">
                    Allows the IdP to verify that SAML requests originated from ShellHub
                  </span>
                </div>
              </label>
            </div>
          )}
        </div>
      </div>
    </Drawer>
  );
}
