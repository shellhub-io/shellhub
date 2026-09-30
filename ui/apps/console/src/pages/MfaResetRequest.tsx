import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { Button, Callout } from "@shellhub/design-system/primitives";
import { useAuthStore } from "../stores/authStore";
import { useMfaResetStore } from "../stores/mfaResetStore";
import AuthActions from "@/components/auth/AuthActions";
import ScreenIntro from "@/components/layout/ScreenIntro";
import { succeeded } from "@/utils/failure";

/**
 * Starts an MFA reset for someone with neither their authenticator nor a recovery code. It goes
 * by mail, so the account's address is the factor being relied on.
 */
export default function MfaResetRequest() {
  const { user, username, mfaToken } = useAuthStore();
  const { requestMfaReset, loading, error } = useMfaResetStore();
  const navigate = useNavigate();

  const identifier = user || username;

  const { handleSubmit } = useForm();

  useEffect(() => {
    useMfaResetStore.setState({ error: null });
  }, []);

  useEffect(() => {
    if (!identifier && !mfaToken) {
      void navigate("/login");
    }
  }, [identifier, mfaToken, navigate]);

  if (!identifier) {
    return null;
  }

  const onSubmit = async () => {
    if (await succeeded(requestMfaReset(identifier))) {
      void navigate("/mfa-reset-verify");
    }
  };

  return (
    <>
      <ScreenIntro
        eyebrow="Two-factor"
        title="Reset MFA via Email"
        lead={
          <>
            We&apos;ll send verification codes to both email addresses
            registered for{" "}
            <span className="font-semibold text-text-primary">
              {identifier}
            </span>
            . You&apos;ll receive two separate emails, and both codes are
            required to complete the reset.
          </>
        }
      />

      <form
        onSubmit={(e) => void handleSubmit(onSubmit)(e)}
        className="space-y-6"
      >
        {error && <Callout variant="error">{error}</Callout>}

        <AuthActions
          primary={
            <Button
              variant="primary"
              size="lg"
              fullWidth
              type="submit"
              loading={loading}
              disabled={loading}
            >
              {loading ? "Sending..." : "Send Verification Codes"}
            </Button>
          }
          links={[{ label: "Back to recovery", to: "/mfa-recover" }]}
        />
      </form>
    </>
  );
}
