import { useState } from "react";
import { useWatch } from "react-hook-form";
import { KeyIcon, ChevronDownIcon, ExclamationCircleIcon } from "@heroicons/react/24/outline";
import { configureSamlAuthentication } from "@/client";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useDrawerForm } from "@/hooks/useDrawerForm";
import { cn } from "@shellhub/design-system/cn";
import FormDrawer from "@/components/common/FormDrawer";
import { FormCheckboxField, FormInputField, FormTextareaField } from "@/components/common/fields/rhf";
import {
  samlSchema,
  buildSamlDefaults,
  buildSamlBody,
  type SamlFormValues,
  type SamlSettings,
} from "./samlSchema";

interface Props {
  open: boolean;
  onClose: () => void;
  onSaved: () => void;
  existingConfig: SamlSettings | null | undefined;
}

export default function SamlConfigDrawer({
  open,
  onClose,
  onSaved,
  existingConfig,
}: Props) {
  const [showAdvanced, setShowAdvanced] = useState(false);

  const form = useDrawerForm(open, samlSchema, buildSamlDefaults(existingConfig));
  const { control, setError, clearErrors } = form;

  const useMetadataUrl = useWatch({ control, name: "useMetadataUrl" });
  const postUrl = useWatch({ control, name: "postUrl" });
  const redirectUrl = useWatch({ control, name: "redirectUrl" });

  const noUrlProvided = !useMetadataUrl && !postUrl && !redirectUrl;

  useResetOnOpen(open, () =>
    setShowAdvanced(
      !!(existingConfig?.idp?.mappings?.email || existingConfig?.idp?.mappings?.name),
    ),
  );

  const onSubmit = async (values: SamlFormValues) => {
    clearErrors("root");
    try {
      await configureSamlAuthentication({
        body: buildSamlBody(values),
        throwOnError: true,
      });
      onSaved();
      onClose();
    } catch {
      setError("root", {
        message:
          "Failed to save SAML configuration. Please check your settings and try again.",
      });
    }
  };

  return (
    <FormDrawer
      form={form}
      onSubmit={onSubmit}
      open={open}
      onClose={onClose}
      title="Configure Single Sign-On"
      subtitle="Configure SAML authentication for your ShellHub instance"
      icon={<KeyIcon className="w-4 h-4 text-primary" />}
      width="md"
      submitLabel="Save Configuration"
    >
      <FormCheckboxField
        name="useMetadataUrl"
        control={control}
        id="saml-use-metadata-url"
        label="Use Metadata URL"
        description="Automatically fetch IdP configuration from a URL"
      />

      {useMetadataUrl ? (
        <FormInputField
          name="metadataUrl"
          control={control}
          id="metadata-url"
          label="IdP Metadata URL"
          type="url"
          placeholder="https://idp.example.com/metadata.xml"
          variant="mono"
        />
      ) : (
        <div className="space-y-4">
          {noUrlProvided && (
            <div className="flex items-start gap-2 bg-accent-yellow/8 border border-accent-yellow/20 text-accent-yellow px-3.5 py-2.5 rounded-md text-xs font-mono">
              <ExclamationCircleIcon
                className="w-3.5 h-3.5 shrink-0 mt-0.5"
                strokeWidth={2}
              />
              Please provide at least one Sign-On URL (POST or Redirect)
            </div>
          )}

          <FormInputField
            name="postUrl"
            control={control}
            id="post-url"
            label="SSO POST URL"
            type="url"
            placeholder="https://idp.example.com/sso/post"
            variant="mono"
          />

          <FormInputField
            name="redirectUrl"
            control={control}
            id="redirect-url"
            label="SSO Redirect URL"
            type="url"
            placeholder="https://idp.example.com/sso/redirect"
            variant="mono"
          />

          <FormInputField
            name="entityId"
            control={control}
            id="entity-id"
            label="Entity ID"
            placeholder="https://idp.example.com/entity"
            variant="mono"
            hint="Issuer/Entity ID from your IdP's SAML configuration"
          />

          <FormTextareaField
            name="certificate"
            control={control}
            id="certificate"
            label="X.509 Certificate"
            rows={5}
            placeholder={
              "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"
            }
            className="resize-none font-mono text-2xs leading-relaxed"
          />
        </div>
      )}

      <div className="border border-border rounded-lg overflow-hidden">
        <button
          type="button"
          onClick={() => setShowAdvanced((v) => !v)}
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
            <FormInputField
              name="emailMapping"
              control={control}
              id="email-mapping"
              label="Email Attribute Mapping"
              placeholder="emailAddress"
              variant="mono"
              hint="SAML attribute name that contains the user's email address"
            />

            <FormInputField
              name="nameMapping"
              control={control}
              id="name-mapping"
              label="Name Attribute Mapping"
              placeholder="displayName"
              variant="mono"
              hint="SAML attribute name that contains the user's display name"
            />

            <FormCheckboxField
              name="signRequests"
              control={control}
              id="saml-sign-requests"
              label="Sign authorization requests"
              description="Allows the IdP to verify that SAML requests originated from ShellHub"
            />
          </div>
        )}
      </div>
    </FormDrawer>
  );
}
