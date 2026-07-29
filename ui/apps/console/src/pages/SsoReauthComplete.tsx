import { useEffect } from "react";

/**
 * Landing page for the SSO re-auth step-up popup. The SAML assertion consumer
 * redirects here with ?status=ok|error after a ForceAuthn re-authentication; it
 * relays the outcome to the opener (WebReauthDialog) and closes itself.
 */
export default function SsoReauthComplete() {
  useEffect(() => {
    const status = new URLSearchParams(window.location.search).get("status");
    const opener = window.opener as Window | null;
    opener?.postMessage(
      status === "ok" ? "sso-reauth-ok" : "sso-reauth-error",
      window.location.origin,
    );
    window.close();
  }, []);

  return (
    <div className="min-h-screen flex items-center justify-center bg-background text-sm text-text-muted">
      Completing re-authentication…
    </div>
  );
}
