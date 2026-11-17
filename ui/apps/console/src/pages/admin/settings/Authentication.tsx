import { useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import {
  KeyIcon,
  ArrowTopRightOnSquareIcon,
  FingerPrintIcon,
} from "@heroicons/react/24/outline";
import {
  configureLocalAuthenticationMutation,
  configureSamlAuthenticationMutation,
  getAuthenticationSettingsOptions,
} from "@/client";
import { isSdkError } from "@/api/errors";
import { adminQueryDefaults } from "@/hooks/useAdminRequests";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";
import SettingsSwitchCard from "@/components/settings/SettingsSwitchCard";
import SettingsField from "@/components/settings/SettingsField";
import PageHeader from "@/components/common/PageHeader";
import { adminNavSectionTitle } from "@/components/layout/adminNav";
import { ADMIN_AUTHENTICATION_PATH } from "@/utils/adminRoute";
import CopyButton from "@/components/common/CopyButton";
import SamlConfigModal from "./SamlConfigModal";
import PageLoader from "@/components/common/PageLoader";
import { Button, Callout, Toggle } from "@shellhub/design-system/primitives";

function toggleErrorMessage(err: unknown, method: string): string {
  return isSdkError(err) && err.status === 400
    ? "You cannot disable all authentication methods."
    : `Failed to update ${method} authentication.`;
}

/**
 * The instance authentication settings: local sign-in and SAML. The API refuses to turn off the
 * last enabled method, and the page shows that refusal beside the toggles.
 */
export default function AdminAuthentication() {
  const settingsQuery = useQuery({
    ...getAuthenticationSettingsOptions(),
    ...adminQueryDefaults,
  });
  const settings = settingsQuery.data;
  const [toggleError, setToggleError] = useState<string | null>(null);
  const [modalOpen, setModalOpen] = useState(false);

  const invalidate = useInvalidateByIds("getAuthenticationSettings", "getInfo");
  const localMutation = useMutation({
    ...configureLocalAuthenticationMutation(),
    onSuccess: invalidate,
  });
  const samlMutation = useMutation({
    ...configureSamlAuthenticationMutation(),
    onSuccess: invalidate,
  });

  const error =
    toggleError ??
    (settingsQuery.isError ? "Failed to load authentication settings." : null);

  const handleLocalToggle = () => {
    setToggleError(null);
    localMutation.mutate(
      { body: { enable: !settings?.local?.enabled } },
      { onError: (err) => setToggleError(toggleErrorMessage(err, "local")) },
    );
  };

  const handleSamlToggle = () => {
    if (!settings?.saml?.enabled) {
      setModalOpen(true);
      return;
    }
    setToggleError(null);
    samlMutation.mutate(
      {
        body: {
          enable: false,
          idp: { entity_id: "", binding: {}, certificate: "" },
          sp: {},
        },
      },
      { onError: (err) => setToggleError(toggleErrorMessage(err, "SAML")) },
    );
  };

  if (settingsQuery.isLoading) {
    return (
      <div>
        <PageHeader
          icon={<KeyIcon className="w-6 h-6" />}
          overline={adminNavSectionTitle(ADMIN_AUTHENTICATION_PATH)}
          title="Authentication"
          description="Control how users authenticate to ShellHub, including local credentials and SAML SSO."
        />
        <div className="mt-8">
          <PageLoader label="Loading settings..." showLabel padding="none" />
        </div>
      </div>
    );
  }

  const localEnabled = settings?.local?.enabled ?? false;
  const samlEnabled = settings?.saml?.enabled ?? false;
  const saml = settings?.saml;

  return (
    <div>
      <PageHeader
        icon={<KeyIcon className="w-6 h-6" />}
        overline={adminNavSectionTitle(ADMIN_AUTHENTICATION_PATH)}
        title="Authentication"
        description="Control how users authenticate to ShellHub, including local credentials and SAML SSO."
      />

      {error && (
        <Callout variant="error" className="mb-5 max-w-2xl">
          {error}
        </Callout>
      )}

      <div className="max-w-2xl space-y-5">
        <SettingsSwitchCard
          icon={<KeyIcon />}
          title="Local authentication"
          description="Users sign in with a username and password stored in ShellHub."
          control={
            <Toggle
              enabled={localEnabled}
              disabled={localMutation.isPending}
              onChange={handleLocalToggle}
              aria-label="Toggle local authentication"
            />
          }
        />

        <SettingsSwitchCard
          icon={<FingerPrintIcon />}
          title="SAML authentication"
          description={
            samlEnabled
              ? "Users sign in through a SAML identity provider (SSO)."
              : "Users sign in through a SAML identity provider (SSO). Turn it on to configure the provider."
          }
          control={
            <Toggle
              enabled={samlEnabled}
              disabled={samlMutation.isPending}
              onChange={handleSamlToggle}
              aria-label="Toggle SAML authentication"
            />
          }
        >
          {samlEnabled && saml && (
            <div className="space-y-4">
              {saml.assertion_url && (
                <SettingsField
                  stacked
                  title="Assertion URL"
                  description="Where the identity provider sends users after they sign in. Set it as the Assertion Consumer Service (ACS) URL in the provider."
                >
                  <div className="flex items-center gap-2">
                    <code className="flex-1 px-3 py-2 bg-background border border-border rounded-lg text-xs font-mono text-text-secondary truncate">
                      {saml.assertion_url}
                    </code>
                    <CopyButton text={saml.assertion_url} size="md" />
                  </div>
                </SettingsField>
              )}

              {saml.idp?.entity_id && (
                <SettingsField
                  stacked
                  title="IdP entity ID"
                  description="How the identity provider names itself."
                >
                  <code className="block px-3 py-2 bg-background border border-border rounded-lg text-xs font-mono text-text-secondary">
                    {saml.idp.entity_id}
                  </code>
                </SettingsField>
              )}

              {saml.idp?.binding?.post && (
                <SettingsField
                  stacked
                  title="IdP sign-on POST URL"
                  description="Where sign-in requests are posted."
                >
                  <code className="block px-3 py-2 bg-background border border-border rounded-lg text-xs font-mono text-text-secondary break-all">
                    {saml.idp.binding.post}
                  </code>
                </SettingsField>
              )}

              {saml.idp?.binding?.redirect && (
                <SettingsField
                  stacked
                  title="IdP sign-on redirect URL"
                  description="Where sign-in requests are redirected."
                >
                  <code className="block px-3 py-2 bg-background border border-border rounded-lg text-xs font-mono text-text-secondary break-all">
                    {saml.idp.binding.redirect}
                  </code>
                </SettingsField>
              )}

              <div className="flex items-center gap-3 pt-1">
                {saml.auth_url && (
                  <Button
                    variant="outline"
                    as="a"
                    size="sm"
                    href={saml.auth_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    title="Opens a new window directly calling the authentication URL"
                    icon={
                      <ArrowTopRightOnSquareIcon
                        className="w-3.5 h-3.5"
                        strokeWidth={2}
                      />
                    }
                  >
                    Test Auth Integration
                  </Button>
                )}

                <Button size="sm" onClick={() => setModalOpen(true)}>
                  Edit Configuration
                </Button>
              </div>
            </div>
          )}
        </SettingsSwitchCard>
      </div>

      <SamlConfigModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        onSaved={() => {
          setToggleError(null);
          void invalidate();
        }}
        existingConfig={saml}
      />
    </div>
  );
}
