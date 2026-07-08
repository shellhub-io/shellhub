import { useState, useEffect, FormEvent } from "react";
import { useForm } from "react-hook-form";
import { isSdkError } from "../api/errors";
import {
  useNavigate,
  Navigate,
  Link,
  useLocation,
  useSearchParams,
} from "react-router-dom";
import {
  LockClosedIcon,
  ArrowRightEndOnRectangleIcon,
} from "@heroicons/react/24/outline";
import { Button, Callout, Spinner } from "@shellhub/design-system/primitives";
import { useAuthStore } from "../stores/authStore";
import { getConfig } from "../env";
import { getSafeRedirect } from "../utils/navigation";
import AuthFooterLinks from "../components/common/AuthFooterLinks";
import { getInfo, getSamlAuthUrl } from "../client";
import {
  FormInputField,
  FormPasswordField,
} from "@/components/common/fields/rhf";
import { loginResolver } from "./setup/loginResolver";
import type { LoginFormValues } from "./setup/loginResolver";

interface CountdownState {
  display: string;
  expired: boolean;
  epoch: number | null;
}

function useLoginCountdown(lockoutEndEpoch: number | null) {
  const [state, setState] = useState<CountdownState>({
    display: "",
    expired: false,
    epoch: null,
  });

  useEffect(() => {
    if (lockoutEndEpoch === null) return;

    const interval = setInterval(() => {
      const diff = lockoutEndEpoch - Date.now() / 1000;
      if (diff <= 0) {
        clearInterval(interval);
        setState({ display: "", expired: true, epoch: lockoutEndEpoch });
      } else if (diff < 60) {
        const s = Math.floor(diff);
        setState({
          display: `${s} ${s === 1 ? "second" : "seconds"}`,
          expired: false,
          epoch: lockoutEndEpoch,
        });
      } else {
        const m = Math.floor(diff / 60);
        setState({
          display: `${m} ${m === 1 ? "minute" : "minutes"}`,
          expired: false,
          epoch: lockoutEndEpoch,
        });
      }
    }, 1000);

    return () => clearInterval(interval);
  }, [lockoutEndEpoch]);

  // If the epoch changed (or was cleared), state is stale — return clean defaults
  if (state.epoch !== lockoutEndEpoch) return { display: "", expired: false };

  return { display: state.display, expired: state.expired };
}

export default function Login() {
  const { cloud: isCloud, enterprise: isEnterprise } = getConfig();
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

  const [error, setError] = useState<string | null>(null);
  const [lockoutEndEpoch, setLockoutEndEpoch] = useState<number | null>(null);
  const { login, loading } = useAuthStore();
  const token = useAuthStore((s) => s.token);
  const navigate = useNavigate();
  const { display: countdownDisplay, expired: lockoutExpired } =
    useLoginCountdown(lockoutEndEpoch);

  const { control, handleSubmit, formState } = useForm<LoginFormValues>({
    resolver: loginResolver,
    mode: "onTouched",
    defaultValues: { username: "", password: "" },
  });

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

  const onSubmit = async (values: LoginFormValues) => {
    setError(null);
    setLockoutEndEpoch(null);
    try {
      await login(values.username, values.password);

      const state = useAuthStore.getState();
      const params = new URLSearchParams(location.search);
      const redirect = getSafeRedirect(params);

      if (state.mfaToken) {
        const mfaPath =
          redirect !== "/dashboard"
            ? `/mfa-login?redirect=${encodeURIComponent(redirect)}`
            : "/mfa-login";
        void navigate(mfaPath);
      } else {
        void navigate(redirect);
      }
    } catch (err) {
      if (!isSdkError(err)) {
        setError("Something went wrong. Please try again later.");
        return;
      }

      switch (err.status) {
        case 401:
          setError(
            "Invalid login credentials. Your password is incorrect or this account doesn't exist.",
          );
          break;
        case 403:
          void navigate(
            `/confirm-account?username=${encodeURIComponent(values.username)}`,
          );
          break;
        case 429: {
          const epoch = Number(err.headers.get("x-account-lockout"));
          setLockoutEndEpoch(isNaN(epoch) ? null : epoch);
          setError(
            "Too many failed login attempts. Please wait before trying again.",
          );
          break;
        }
        default:
          setError("Something went wrong on our end. Please try again later.");
      }
    }
  };

  const handleFormSubmit = (e: FormEvent) => {
    void handleSubmit(onSubmit)(e);
  };

  // On enterprise, show the local form only once we know local auth is enabled.
  // Using an explicit === true guard (not !ssoOnly) prevents the form from
  // flashing while authentication info is still loading (null state).
  const showLocalForm = !isEnterprise || authentication?.local === true;
  const ssoOnly = isEnterprise && authentication?.local === false;

  // Already authenticated (e.g. straight after setup's auto-login): the login form
  // has nothing to do here, so send the user into the app. Skipped while a query
  // token is being exchanged, which deliberately re-authenticates.
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
    <div className="w-full max-w-5xl mx-auto px-8 py-12 flex flex-col items-center">
      {/* Hero */}
      <div className="text-center mb-12 animate-fade-in">
        <div className="animate-float mb-6 inline-block">
          <div className="w-20 h-20 rounded-2xl bg-primary/15 border border-primary/25 flex items-center justify-center shadow-lg shadow-primary/10">
            <LockClosedIcon
              className="w-10 h-10 text-primary"
              strokeWidth={1.2}
            />
          </div>
        </div>

        <p className="text-2xs font-mono font-semibold uppercase tracking-wide text-primary/80 mb-2">
          Welcome Back
        </p>
        <h1 className="text-3xl font-bold text-text-primary mb-3">
          Sign in to ShellHub
        </h1>
        <p className="text-sm text-text-muted max-w-md mx-auto leading-relaxed">
          Access your devices, sessions, and security rules from a single
          dashboard.
        </p>
      </div>

      {/* Alerts — rendered outside the form so they are visible in SSO-only mode too */}
      {(lockoutExpired ||
        !!notice ||
        !!missingAssertions ||
        (!!error && !lockoutExpired)) && (
        <div className="w-full max-w-sm flex flex-col gap-3 mb-4">
          {lockoutExpired && (
            <Callout variant="success">
              Your timeout has finished. Please try to log back in.
            </Callout>
          )}
          {notice && <Callout variant="success">{notice}</Callout>}
          {missingAssertions && (
            <Callout variant="error">
              The SSO configuration is incomplete due to missing required
              mappings. Please contact your administrator.
            </Callout>
          )}
          {error && !lockoutExpired && (
            <Callout variant="error">
              <span>
                {error}
                {countdownDisplay && (
                  <span className="font-semibold"> ({countdownDisplay})</span>
                )}
              </span>
            </Callout>
          )}
        </div>
      )}

      {/* Form card — only shown once we know local auth is enabled */}
      {showLocalForm && (
        <div
          className="w-full max-w-sm bg-card/80 border border-border rounded-2xl p-8 backdrop-blur-sm animate-slide-up"
          style={{ animationDelay: "200ms" }}
        >
          <form onSubmit={handleFormSubmit} className="space-y-5">
            <FormInputField<LoginFormValues>
              id="username"
              label="Username"
              name="username"
              control={control}
              placeholder="username"
              autoComplete="username"
            />

            <FormPasswordField<LoginFormValues>
              id="password"
              label="Password"
              name="password"
              control={control}
              placeholder="password"
              autoComplete="current-password"
            />

            {isCloud && (
              <div className="flex justify-end">
                <Link
                  to="/forgot-password"
                  className="text-2xs text-text-muted hover:text-text-secondary transition-colors"
                >
                  Forgot password?
                </Link>
              </div>
            )}

            <Button
              variant="primary"
              size="lg"
              fullWidth
              type="submit"
              className="px-4"
              loading={loading}
              disabled={!formState.isValid || loading}
            >
              {loading ? "Authenticating..." : "Sign In"}
            </Button>
          </form>
        </div>
      )}

      {/* SSO login */}
      {isEnterprise && authentication?.saml && (
        <div
          className="w-full max-w-sm animate-slide-up"
          style={{ animationDelay: ssoOnly ? "200ms" : "300ms" }}
        >
          {!ssoOnly && (
            <div className="flex items-center gap-3 my-4">
              <div className="flex-1 h-px bg-border" />
              <span className="text-2xs font-mono text-text-muted uppercase tracking-label">
                or
              </span>
              <div className="flex-1 h-px bg-border" />
            </div>
          )}

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
        </div>
      )}

      {/* Footer links */}
      <AuthFooterLinks />
    </div>
  );
}
