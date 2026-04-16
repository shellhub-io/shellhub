import { useState, useEffect, useCallback } from "react";
import {
  KeyIcon,
  ExclamationCircleIcon,
  ArrowTopRightOnSquareIcon,
} from "@heroicons/react/24/outline";
import {
  getAuthenticationSettings,
  configureLocalAuthentication,
  configureSamlAuthentication,
} from "../../../client";
import type { GetAuthenticationSettingsResponse } from "../../../client";
import { isSdkError } from "../../../api/errors";
import PageHeader from "../../../components/common/PageHeader";
import CopyButton from "../../../components/common/CopyButton";
import SamlConfigDrawer from "./SamlConfigDrawer";

type AuthSettings = GetAuthenticationSettingsResponse;

function Toggle({
  enabled,
  loading,
  onToggle,
  ariaLabel,
}: {
  enabled: boolean;
  loading: boolean;
  onToggle: () => void;
  ariaLabel: string;
}) {
  return (
    <button
      role="switch"
      type="button"
      aria-checked={enabled}
      aria-label={ariaLabel}
      disabled={loading}
      onClick={onToggle}
      className={`relative inline-flex h-5 w-9 shrink-0 items-center rounded-full transition-colors duration-200 disabled:opacity-dim disabled:cursor-not-allowed focus:outline-none focus:ring-2 focus:ring-primary/30 focus:ring-offset-2 focus:ring-offset-card ${
        enabled ? "bg-primary" : "bg-border"
      }`}
    >
      <span
        className={`inline-block h-3.5 w-3.5 rounded-full bg-white shadow-sm transition-transform duration-200 ${
          enabled ? "translate-x-[18px]" : "translate-x-[2px]"
        }`}
      />
    </button>
  );
}

export default function AdminAuthentication() {
  const [settings, setSettings] = useState<AuthSettings | null>(null);
  const [loading, setLoading] = useState(true);
  const [togglingLocal, setTogglingLocal] = useState(false);
  const [togglingSaml, setTogglingSaml] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);

  const refresh = useCallback(() => {
    setRefreshKey((k) => k + 1);
  }, []);

  useEffect(() => {
    let cancelled = false;
    void (async () => {
      try {
        const { data } = await getAuthenticationSettings({ throwOnError: true });
        if (!cancelled) setSettings(data);
      } catch {
        if (!cancelled) setError("Failed to load authentication settings.");
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => { cancelled = true; };
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
      // Turning ON — open drawer to configure first
      setDrawerOpen(true);
      return;
    }
    // Turning OFF
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
          overline="Admin Settings"
          title="Authentication"
          description="Control how users authenticate to ShellHub, including local credentials and SAML SSO."
        />
        <div className="flex items-center gap-3 mt-8" role="status">
          <span className="w-4 h-4 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
          <span className="text-xs font-mono text-text-muted">Loading settings...</span>
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
        overline="Admin Settings"
        title="Authentication"
        description="Control how users authenticate to ShellHub, including local credentials and SAML SSO."
      />

      {error && (
        <div
          role="alert"
          className="flex items-center gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3.5 py-2.5 rounded-md text-xs font-mono mb-5 animate-slide-down"
        >
          <ExclamationCircleIcon className="w-3.5 h-3.5 shrink-0" strokeWidth={2} />
          {error}
        </div>
      )}

      <div className="space-y-4">
        {/* Local Authentication */}
        <div className="bg-card border border-border rounded-xl p-5">
          <div className="flex items-center justify-between gap-4">
            <div className="flex-1">
              <h3 className="text-sm font-semibold text-text-primary mb-0.5">
                Local Authentication
              </h3>
              <p className="text-xs text-text-muted leading-relaxed">
                Allow users to sign in with a username and password stored locally.
              </p>
            </div>
            <Toggle
              enabled={localEnabled}
              loading={togglingLocal}
              onToggle={() => void handleLocalToggle()}
              ariaLabel="Toggle local authentication"
            />
          </div>
        </div>

        {/* SAML Authentication */}
        <div className="bg-card border border-border rounded-xl overflow-hidden">
          <div className="flex items-center justify-between gap-4 p-5">
            <div className="flex-1">
              <h3 className="text-sm font-semibold text-text-primary mb-0.5">
                SAML Authentication
              </h3>
              <p className="text-xs text-text-muted leading-relaxed">
                Allow users to sign in via a SAML Identity Provider (SSO).
                {!samlEnabled && (
                  <span className="text-text-secondary">
                    {" "}Enable to configure your IdP settings.
                  </span>
                )}
              </p>
            </div>
            <Toggle
              enabled={samlEnabled}
              loading={togglingSaml}
              onToggle={() => void handleSamlToggle()}
              ariaLabel="Toggle SAML authentication"
            />
          </div>

          {/* SSO Details — shown only when SAML is enabled */}
          {samlEnabled && saml && (
            <div className="border-t border-border px-5 pb-5 pt-4 space-y-4">
              <h4 className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted">
                SSO Configuration
              </h4>

              {/* Assertion URL */}
              {saml.assertion_url && (
                <div>
                  <p className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-1.5">
                    Assertion URL
                  </p>
                  <div className="flex items-center gap-2">
                    <code className="flex-1 px-3 py-2 bg-background border border-border rounded-lg text-xs font-mono text-text-secondary truncate">
                      {saml.assertion_url}
                    </code>
                    <CopyButton text={saml.assertion_url} size="md" />
                  </div>
                  <p className="mt-1 text-2xs text-text-muted leading-relaxed">
                    The URL where your IdP should redirect users after successful authentication.
                    Configure this as the Assertion Consumer Service (ACS) URL in your IdP.
                  </p>
                </div>
              )}

              {/* IdP Entity ID */}
              {saml.idp?.entity_id && (
                <div>
                  <p className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-1.5">
                    IdP Entity ID
                  </p>
                  <code className="block px-3 py-2 bg-background border border-border rounded-lg text-xs font-mono text-text-secondary">
                    {saml.idp.entity_id}
                  </code>
                </div>
              )}

              {/* Binding URLs */}
              {saml.idp?.binding?.post && (
                <div>
                  <p className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-1.5">
                    IdP SignOn POST URL
                  </p>
                  <code className="block px-3 py-2 bg-background border border-border rounded-lg text-xs font-mono text-text-secondary break-all">
                    {saml.idp.binding.post}
                  </code>
                </div>
              )}

              {saml.idp?.binding?.redirect && (
                <div>
                  <p className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-1.5">
                    IdP SignOn Redirect URL
                  </p>
                  <code className="block px-3 py-2 bg-background border border-border rounded-lg text-xs font-mono text-text-secondary break-all">
                    {saml.idp.binding.redirect}
                  </code>
                </div>
              )}

              {/* Actions */}
              <div className="flex items-center gap-3 pt-1">
                {saml.auth_url && (
                  <a
                    href={saml.auth_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-text-secondary border border-border hover:border-border-light hover:text-text-primary rounded-lg transition-all"
                    title="Opens a new window directly calling the authentication URL"
                  >
                    <ArrowTopRightOnSquareIcon className="w-3.5 h-3.5" strokeWidth={2} />
                    Test Auth Integration
                  </a>
                )}

                <button
                  type="button"
                  onClick={() => setDrawerOpen(true)}
                  className="flex items-center gap-1.5 px-3 py-1.5 bg-primary hover:bg-primary-600 text-white text-xs font-semibold rounded-lg transition-all"
                >
                  Edit Configuration
                </button>
              </div>
            </div>
          )}
        </div>
      </div>

      <SamlConfigDrawer
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        onSaved={() => void handleSamlSaved()}
        existingConfig={saml}
      />
    </div>
  );
}
