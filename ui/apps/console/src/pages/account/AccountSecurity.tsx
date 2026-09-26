import { useState } from "react";
import { useWatch } from "react-hook-form";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useDrawerForm } from "@/hooks/useDrawerForm";
import { useAuthStore } from "@/stores/authStore";
import FormModal from "@/components/common/FormModal";
import { isSdkError } from "@/api/errors";

import {
  changePasswordSchema,
  type ChangePasswordFormValues,
} from "@/pages/account/changePasswordSchema";
import {
  FormPasswordField,
} from "@/components/common/fields/rhf";
import {
  CheckIcon,
  LockClosedIcon,
  ShieldCheckIcon,
} from "@heroicons/react/24/outline";
import MfaEnableModal from "@/components/mfa/MfaEnableModal";
import MfaDisableDialog from "@/components/mfa/MfaDisableDialog";
import { isEnterpriseOrCloud } from "@/env";
import { Button } from "@shellhub/design-system/primitives";
import SettingsSection from "@/components/settings/SettingsSection";
import SettingsField from "@/components/settings/SettingsField";
import SettingsSwitchCard from "@/components/settings/SettingsSwitchCard";

function ChangePasswordModal({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const updatePw = useAuthStore((s) => s.updatePassword);
  const [success, setSuccess] = useState(false);

  const form = useDrawerForm(open, changePasswordSchema, {
    current: "",
    newPw: "",
    confirmPw: "",
  });
  const { control, setError, clearErrors } = form;

  const [watchedCurrent, watchedNewPw, watchedConfirmPw] = useWatch({
    control,
    name: ["current", "newPw", "confirmPw"],
  });
  const allFilled = !!watchedCurrent && !!watchedNewPw && !!watchedConfirmPw;

  useResetOnOpen(open, () => setSuccess(false));

  const onValid = async (values: ChangePasswordFormValues) => {
    clearErrors("root");
    try {
      await updatePw(values.current, values.newPw);
      setSuccess(true);
      setTimeout(onClose, 1200);
    } catch (err) {
      const errorMessage =
        isSdkError(err) && err.status === 403
          ? "Current password is incorrect."
          : "Failed to change password.";

      setError("root", { message: errorMessage });
    }
  };

  return (
    <FormModal
      size="sm"
      form={form}
      onSubmit={onValid}
      open={open}
      onClose={onClose}
      icon={<LockClosedIcon />}
      title="Change password"
      description="Enter your current password, then the one to replace it."
      submitLabel="Change password"
      submitDisabled={!allFilled}
      submitIcon={<CheckIcon className="w-4 h-4" strokeWidth={2} />}
    >
      <FormPasswordField
        name="current"
        control={control}
        id="change-pw-current"
        label="Current Password"
        autoComplete="current-password"
      />
      <FormPasswordField
        name="newPw"
        control={control}
        id="change-pw-new"
        label="New Password"
        autoComplete="new-password"
        hint="5-32 characters"
      />
      <FormPasswordField
        name="confirmPw"
        control={control}
        id="change-pw-confirm"
        label="Confirm New Password"
        autoComplete="new-password"
      />
      {success && (
        <p className="text-2xs text-accent-green">
          Password changed successfully.
        </p>
      )}
    </FormModal>
  );
}

/**
 * The security section of the account: the password and multi-factor authentication, or a note
 * that the identity provider manages both for an SSO user.
 */
export default function AccountSecurity() {
  const { origin, mfaEnabled, recoveryEmail, fetchUser } = useAuthStore();
  const isSsoUser = origin === "saml";
  const [pwModalOpen, setPwModalOpen] = useState(false);
  const [mfaEnableOpen, setMfaEnableOpen] = useState(false);
  const [mfaDisableOpen, setMfaDisableOpen] = useState(false);

  return (
    <>
      <SettingsSection
        title="Security"
        description="How you prove it's you when you sign in."
      >
        {isSsoUser ? (
          <SettingsField
            title="Managed by your identity provider"
            description="You sign in through SSO, so your password and multi-factor authentication are set with your identity provider, not in ShellHub."
          >
            <span className="px-1.5 py-0.5 text-2xs font-mono font-semibold uppercase tracking-wider rounded bg-primary/10 text-primary border border-primary/20">
              SSO
            </span>
          </SettingsField>
        ) : (
          <>
            <SettingsField
              title="Password"
              description="The password you sign in with."
            >
              <Button
                size="sm"
                variant="secondary"
                onClick={() => setPwModalOpen(true)}
              >
                Change Password
              </Button>
            </SettingsField>

            <SettingsSwitchCard
              icon={<ShieldCheckIcon />}
              title="Multi-factor authentication"
              description={
                isEnterpriseOrCloud()
                  ? "A code from an authenticator app on every sign-in. Recovery codes are shown once, during setup."
                  : "A code from an authenticator app on every sign-in. Available on paid editions."
              }
              control={
                !isEnterpriseOrCloud() ? (
                  <Button
                    as="a"
                    size="sm"
                    variant="secondary"
                    href="https://www.shellhub.io/pricing"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    Upgrade
                  </Button>
                ) : mfaEnabled ? (
                  <div className="flex items-center gap-3">
                    <span className="px-1.5 py-0.5 text-2xs font-mono font-semibold uppercase tracking-wider rounded bg-accent-green/10 text-accent-green border border-accent-green/20">
                      Enabled
                    </span>
                    <Button
                      size="sm"
                      variant="secondary"
                      onClick={() => setMfaDisableOpen(true)}
                    >
                      Disable
                    </Button>
                  </div>
                ) : (
                  <Button
                    size="sm"
                    variant="secondary"
                    onClick={() => setMfaEnableOpen(true)}
                  >
                    Enable MFA
                  </Button>
                )
              }
            />
          </>
        )}
      </SettingsSection>
      <ChangePasswordModal
        open={pwModalOpen}
        onClose={() => setPwModalOpen(false)}
      />
      <MfaEnableModal
        open={mfaEnableOpen}
        onClose={() => setMfaEnableOpen(false)}
        onSuccess={() => {
          setMfaEnableOpen(false);
          void fetchUser();
        }}
        currentRecoveryEmail={recoveryEmail ?? null}
      />
      <MfaDisableDialog
        open={mfaDisableOpen}
        onClose={() => setMfaDisableOpen(false)}
        onSuccess={() => {
          setMfaDisableOpen(false);
          void fetchUser();
        }}
      />
    </>
  );
}
