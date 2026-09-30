import { useState, useEffect } from "react";
import {
  useNavigate,
  Navigate,
  useLocation,
  useSearchParams,
} from "react-router-dom";
import { ArrowRightEndOnRectangleIcon } from "@heroicons/react/24/outline";
import { Button, Callout, Spinner } from "@shellhub/design-system/primitives";
import { useAuthStore } from "../stores/authStore";
import { isCloud, isEnterpriseOrCloud } from "../env";
import { getSafeRedirect } from "@/utils/navigation";
import PendingDeviceCallout from "@/components/auth/PendingDeviceCallout";
import AuthActions from "@/components/auth/AuthActions";
import SignInForm from "@/components/auth/SignInForm";
import ScreenIntro from "@/components/layout/ScreenIntro";
import { getInfo, getSamlAuthUrl } from "../client";

/**
 * The sign-in page. What it offers depends on the edition — SSO and sign-up exist only above
 * community — and on where the user was going, which it returns them to afterwards.
 */
export default function Login() {
  const isCloudEdition = isCloud();
  const isEnterpriseOrCloudEdition = isEnterpriseOrCloud();
  const location = useLocation();
  const rawState = location.state as Record<string, unknown> | null;
  const notice =
    typeof rawState?.notice === "string" ? rawState.notice : undefined;

  const [searchParams] = useSearchParams();
  const queryToken = searchParams.get("token");
  const missingAssertions = searchParams.get("missing_assertions");
  const [tokenLoading, setTokenLoading] = useState(!!queryToken);
  const [authentication, setAuthentication] = useState<{
    local?: boolean;
    saml?: boolean;
  } | null>(null);
  const [ssoLoading, setSsoLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (notice) {
      window.history.replaceState({}, document.title);
    }
  }, [notice]);

  useEffect(() => {
    void getInfo()
      .then(({ data }) => setAuthentication(data?.authentication ?? null))
      .catch(() => setAuthentication(null));
  }, []);

  const token = useAuthStore((s) => s.token);
  const navigate = useNavigate();

  useEffect(() => {
    if (!queryToken) return;

    const { logout, loginWithToken } = useAuthStore.getState();
    logout();

    loginWithToken(queryToken)
      .then(() => navigate("/dashboard"))
      .catch(() => {
        setTokenLoading(false);
        setError("Failed to authenticate with the provided token.");
      });
  }, [queryToken, navigate]);

  const handleSsoLogin = async () => {
    setSsoLoading(true);
    try {
      const { data } = await getSamlAuthUrl({ throwOnError: true });
      window.location.replace(data.url);
    } catch {
      setError("Failed to retrieve SSO login URL. Please try again.");
      setSsoLoading(false);
    }
  };

  const showLocalForm =
    !isEnterpriseOrCloudEdition || authentication?.local === true;
  const ssoOnly = isEnterpriseOrCloudEdition && authentication?.local === false;

  if (token && !queryToken) {
    return <Navigate to="/" replace />;
  }

  if (tokenLoading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <Spinner size="xl" />
      </div>
    );
  }

  return (
    <>
      <ScreenIntro
        eyebrow="Sign in"
        title="Sign in to ShellHub"
        lead="Access your devices, sessions, and security rules from a single dashboard."
      />

      <div className="flex flex-col gap-3 mb-6 empty:hidden">
        <PendingDeviceCallout />
        {notice && <Callout variant="success">{notice}</Callout>}
        {missingAssertions && (
          <Callout variant="error">
            The SSO configuration is incomplete due to missing required
            mappings. Please contact your administrator.
          </Callout>
        )}
        {error && <Callout variant="error">{error}</Callout>}
      </div>

      {isEnterpriseOrCloudEdition && authentication?.saml && (
        <div>
          <Button
            variant={ssoOnly ? "primary" : "secondary"}
            fullWidth
            loading={ssoLoading}
            disabled={ssoLoading}
            icon={<ArrowRightEndOnRectangleIcon className="w-4 h-4" />}
            data-testid="sso-btn"
            onClick={() => void handleSsoLogin()}
          >
            Login with SSO
          </Button>

          {!ssoOnly && (
            <div className="flex items-center gap-3 my-5">
              <div className="flex-1 h-px bg-border" />
              <span className="text-2xs font-mono text-text-muted uppercase tracking-label">
                or
              </span>
              <div className="flex-1 h-px bg-border" />
            </div>
          )}
        </div>
      )}

      {showLocalForm && (
        <SignInForm
          redirect={getSafeRedirect(new URLSearchParams(location.search))}
        />
      )}

      {isCloudEdition && (
        <div className="mt-4">
          <AuthActions
            links={[
              ...(showLocalForm
                ? [{ label: "Forgot password?", to: "/forgot-password" }]
                : []),
              { label: "Don't have an account? Sign up", to: "/sign-up" },
            ]}
          />
        </div>
      )}
    </>
  );
}
