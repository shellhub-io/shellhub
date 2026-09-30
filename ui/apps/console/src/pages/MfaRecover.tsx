import { useState, FormEvent, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { Button, Callout } from "@shellhub/design-system/primitives";
import { useAuthStore } from "@/stores/authStore";
import { recoveryDisableMfa } from "@/client";
import MfaRecoveryTimeoutModal from "@/components/mfa/MfaRecoveryTimeoutModal";
import { landingAfterSignIn } from "@/utils/navigation";
import AuthActions from "@/components/auth/AuthActions";
import ScreenIntro from "@/components/layout/ScreenIntro";
import { FormInputField } from "@/components/common/fields/rhf";
import { mfaRecoverResolver } from "./setup/mfaRecoverResolver";
import type { MfaRecoverFormValues } from "./setup/mfaRecoverResolver";

/**
 * Signing in with a recovery code, for someone without their authenticator. Each code works
 * once, which the page says, because using one silently reduces what is left.
 */
export default function MfaRecover() {
  const {
    recoverWithCode,
    loading,
    error,
    mfaRecoveryExpiry,
    updateMfaStatus,
    user,
    username,
    mfaToken,
  } = useAuthStore();
  const navigate = useNavigate();

  const identifier = user || username;
  const [showTimeoutModal, setShowTimeoutModal] = useState(false);

  const { control, handleSubmit, reset, resetField, formState } =
    useForm<MfaRecoverFormValues>({
      resolver: mfaRecoverResolver,
      mode: "onTouched",
      defaultValues: { recoveryCode: "" },
    });

  useEffect(() => {
    useAuthStore.setState({ error: null });
  }, []);

  useEffect(() => {
    if (!identifier && !mfaToken) {
      void navigate("/login");
    }
  }, [identifier, mfaToken, navigate]);

  if (!identifier) {
    return null;
  }

  const onSubmit = async (values: MfaRecoverFormValues) => {
    try {
      await recoverWithCode(values.recoveryCode, identifier);
      reset();
      setShowTimeoutModal(true);
    } catch {
      resetField("recoveryCode");
    }
  };

  const handleFormSubmit = (e: FormEvent) => {
    void handleSubmit(onSubmit)(e);
  };

  const handleDisableMfa = async () => {
    await recoveryDisableMfa({ throwOnError: true });
    updateMfaStatus(false);
    setShowTimeoutModal(false);
    void navigate(landingAfterSignIn("/dashboard"));
  };

  const handleCloseModal = () => {
    setShowTimeoutModal(false);
    useAuthStore.setState({ mfaRecoveryExpiry: null });
    void navigate(landingAfterSignIn("/dashboard"));
  };

  return (
    <>
      <ScreenIntro
        eyebrow="Two-factor"
        title="Recover Your Account"
        lead={
          <>
            Enter one of your recovery codes for{" "}
            <span className="font-semibold text-text-primary">
              {identifier}
            </span>
            . Each code works once. After using one, you have a 10-minute window
            to disable MFA if you no longer have your authenticator.
          </>
        }
      />

      <form onSubmit={handleFormSubmit} className="space-y-6">
        {error && <Callout variant="error">{error}</Callout>}

        <FormInputField<MfaRecoverFormValues>
          id="recovery-code"
          label="Recovery Code"
          name="recoveryCode"
          control={control}
          variant="mono"
          placeholder="Enter recovery code"
          hint="You received 6 recovery codes when you enabled MFA."
        />

        <AuthActions
          primary={
            <Button
              variant="warning"
              size="lg"
              fullWidth
              type="submit"
              loading={loading}
              disabled={loading || !formState.isValid}
            >
              {loading ? "Recovering..." : "Recover Account"}
            </Button>
          }
          links={[
            { label: "Back to verification", to: "/mfa-login" },
            {
              label: "Lost the codes? Reset by email",
              to: "/mfa-reset-request",
            },
          ]}
        />
      </form>

      {showTimeoutModal && mfaRecoveryExpiry && (
        <MfaRecoveryTimeoutModal
          open={showTimeoutModal}
          expiresAt={mfaRecoveryExpiry}
          onClose={handleCloseModal}
          onDisable={handleDisableMfa}
        />
      )}
    </>
  );
}
