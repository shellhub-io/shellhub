import { useState, useEffect, useEffectEvent } from "react";
import {
  useNavigate,
  Navigate,
  useLocation,
  useSearchParams,
} from "react-router-dom";
import { ArrowRightEndOnRectangleIcon } from "@heroicons/react/24/outline";
import { Button, Callout, Spinner } from "@shellhub/design-system/primitives";
import { useAuthStore } from "../stores/authStore";
import { isCloud, isEnterprise } from "../env";
import { getSafeRedirect } from "@/utils/navigation";
import PendingDeviceCallout from "@/components/auth/PendingDeviceCallout";
import AuthActions from "@/components/auth/AuthActions";
import SignInForm from "@/components/auth/SignInForm";
import ScreenIntro from "@/components/layout/ScreenIntro";
import { isTokenExpired } from "@/utils/authToken";
import { getInfo, getSamlAuthUrl } from "../client";

const SSO_RESTART_KEY = "shellhub.ssoRestart";
const SSO_RESTART_WINDOW_MS = 2 * 60_000;

function restartedRecently() {
  const restartedAt = Number(sessionStorage.getItem(SSO_RESTART_KEY));
  return restartedAt > 0 && Date.now() - restartedAt < SSO_RESTART_WINDOW_MS;
}

async function startSsoSignIn() {
  const { data } = await getSamlAuthUrl({ throwOnError: true });
  window.location.replace(data.url);
}

/**
 * The sign-in page. What it offers depends on the edition (SSO exists only on enterprise, sign-up
 * only on cloud) and on where the user was going, which it returns them to afterwards. Reached with
 * `signOut: true` in the location state, it ends the current session itself and keeps the notice
 * beside it. Reached with `?sso=start`, where the server sends a sign-in the IdP started on its
 * own, it starts an SSO sign-in, and shows an error instead if it started one in this tab within
 * the last two minutes, which means the IdP keeps sending sign-ins it was not asked for. A session
 * whose token expired counts as signed out here, so it neither sends the page away nor blocks the
 * sign-in.
 */
export default function Login() {
  const isCloudEdition = isCloud();
  const isEnterpriseEdition = isEnterprise();
  const location = useLocation();
  const { notice, signOut } = (location.state ?? {}) as {
    notice?: string;
    signOut?: boolean;
  };

  const [searchParams] = useSearchParams();
  const queryToken = searchParams.get("token");
  const missingAssertions = searchParams.get("missing_assertions");
  const token = useAuthStore((s) => s.token);
  const signedIn = token !== null && !isTokenExpired(token);
  const ssoRestart = searchParams.get("sso") === "start" && !signedIn;
  const [ssoRestartRefused] = useState(
    () => ssoRestart && restartedRecently(),
  );
  const [tokenLoading, setTokenLoading] = useState(!!queryToken);
  const [authentication, setAuthentication] = useState<{
    local?: boolean;
    saml?: boolean;
  } | null>(null);
  const [ssoLoading, setSsoLoading] = useState(
    ssoRestart && !ssoRestartRefused,
  );
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const { token: stored, logout } = useAuthStore.getState();
    if (stored && isTokenExpired(stored)) {
      logout();
    }
  }, []);

  useEffect(() => {
    if (notice) {
      window.history.replaceState({}, document.title);
    }
  }, [notice]);

  useEffect(() => {
    if (signOut) {
      useAuthStore.getState().logout();
    }
  }, [signOut]);

  useEffect(() => {
    void getInfo()
      .then(({ data }) => setAuthentication(data?.authentication ?? null))
      .catch(() => setAuthentication(null));
  }, []);

  const navigate = useNavigate();

  useEffect(() => {
    if (!queryToken) return;

    const { logout, loginWithToken } = useAuthStore.getState();
    logout();

    loginWithToken(queryToken)
      .then(() => {
        sessionStorage.removeItem(SSO_RESTART_KEY);
        void navigate("/dashboard");
      })
      .catch(() => {
        setTokenLoading(false);
        setError("Failed to authenticate with the provided token.");
      });
  }, [queryToken, navigate]);

  const failSsoStart = () => {
    sessionStorage.removeItem(SSO_RESTART_KEY);
    setError("Failed to retrieve SSO login URL. Please try again.");
    setSsoLoading(false);
  };
  const onRestartFailed = useEffectEvent(failSsoStart);

  useEffect(() => {
    if (!ssoRestart) return;

    if (ssoRestartRefused) {
      sessionStorage.removeItem(SSO_RESTART_KEY);
      return;
    }

    sessionStorage.setItem(SSO_RESTART_KEY, String(Date.now()));
    startSsoSignIn().catch(() => onRestartFailed());
  }, [ssoRestart, ssoRestartRefused]);

  const handleSsoLogin = () => {
    setSsoLoading(true);
    startSsoSignIn().catch(failSsoStart);
  };

  const showLocalForm = !isEnterpriseEdition || authentication?.local === true;
  const ssoOnly = isEnterpriseEdition && authentication?.local === false;

  if (signedIn && !queryToken && !signOut) {
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
        {ssoRestartRefused && (
          <Callout variant="error">
            Your identity provider keeps starting sign-ins ShellHub did not
            request. Use the SSO button to sign in.
          </Callout>
        )}
        {error && <Callout variant="error">{error}</Callout>}
      </div>

      {isEnterpriseEdition && authentication?.saml && (
        <div>
          <Button
            variant={ssoOnly ? "primary" : "secondary"}
            fullWidth
            loading={ssoLoading}
            disabled={ssoLoading}
            icon={<ArrowRightEndOnRectangleIcon className="w-4 h-4" />}
            data-testid="sso-btn"
            onClick={handleSsoLogin}
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
              { label: "Forgot password?", to: "/forgot-password" },
              { label: "Don't have an account? Sign up", to: "/sign-up" },
            ]}
          />
        </div>
      )}
    </>
  );
}
