import { useState, useEffect, useCallback } from "react";
import {
  KeyIcon,
  ArrowTopRightOnSquareIcon,
  FingerPrintIcon,
} from "@heroicons/react/24/outline";
import {
  getAuthenticationSettings,
  configureLocalAuthentication,
  configureSamlAuthentication,
} from "@/client";
import type { GetAuthenticationSettingsResponse } from "@/client";
import { isSdkError } from "@/api/errors";
import SettingsSwitchCard from "@/components/settings/SettingsSwitchCard";
import SettingsField from "@/components/settings/SettingsField";
import PageHeader from "@/components/common/PageHeader";
import { adminNavSectionTitle } from "@/components/layout/adminNav";
import CopyButton from "@/components/common/CopyButton";
import SamlConfigModal from "./SamlConfigModal";
import PageLoader from "@/components/common/PageLoader";
import { Button, Callout, Toggle } from "@shellhub/design-system/primitives";

type AuthSettings = GetAuthenticationSettingsResponse;

/**
 * The instance authentication settings: local sign-in and SAML. Turning local sign-in off with
 * no working SSO would lock everyone out, so the page guards that rather than the API alone.
 */
export default function AdminAuthentication() {
  const [settings, setSettings] = useState<AuthSettings | null>(null);
  const [loading, setLoading] = useState(true);
  const [togglingLocal, setTogglingLocal] = useState(false);
  const [togglingSaml, setTogglingSaml] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);

  const refresh = useCallback(() => {
    setRefreshKey((k) => k + 1);
  }, []);

  useEffect(() => {
    let cancelled = false;
    void (async () => {
      try {
        const { data } = await getAuthenticationSettings({
          throwOnError: true,
        });
        if (!cancelled) setSettings(data);
      } catch {
        if (!cancelled) setError("Failed to load authentication settings.");
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [refreshKey]);

  const handleLocalToggle = async () => {
    setTogglingLocal(true);
    setError(null);
    try {
      await configureLocalAuthentication({
        body: { enable: !settings?.local?.enabled },
        throwOnError: true,
      });
      refresh();
    } catch (err) {
      setError(
        isSdkError(err) && err.status === 400
          ? "You cannot disable all authentication methods."
          : "Failed to update local authentication.",
      );
    } finally {
      setTogglingLocal(false);
    }
  };

  const handleSamlToggle = async () => {
    if (!settings?.saml?.enabled) {
      setModalOpen(true);
      return;
    }
    setTogglingSaml(true);
    setError(null);
    try {
      await configureSamlAuthentication({
        body: {
          enable: false,
          idp: { entity_id: "", binding: {}, certificate: "" },
          sp: {},
        },
        throwOnError: true,
      });
      refresh();
    } catch (err) {
      setError(
        isSdkError(err) && err.status === 400
          ? "You cannot disable all authentication methods."
          : "Failed to update SAML authentication.",
      );
    } finally {
      setTogglingSaml(false);
    }
  };

  const handleSamlSaved = () => {
    refresh();
  };

  if (loading) {
    return (
      <div>
        <PageHeader
          icon={<KeyIcon className="w-6 h-6" />}
          overline={adminNavSectionTitle("/admin/settings/authentication")}
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
        overline={adminNavSectionTitle("/admin/settings/authentication")}
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
              disabled={togglingLocal}
              onChange={() => void handleLocalToggle()}
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
              disabled={togglingSaml}
              onChange={() => void handleSamlToggle()}
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
        onSaved={() => void handleSamlSaved()}
        existingConfig={saml}
      />
    </div>
  );
}
