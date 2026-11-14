import { useState } from "react";
import { useWatch } from "react-hook-form";
import { useDrawerForm } from "@/hooks/useDrawerForm";
import { useAuthStore } from "@/stores/authStore";
import FormModal from "@/components/common/FormModal";
import { apiErrorFields, isSdkError } from "@/api/errors";
import { emailRefusalMessage } from "@/utils/emailRefusal";
import {
  changesAnAddress,
  editProfileSchema,
  type EditProfileFormValues,
  type CurrentProfileValues,
} from "@/pages/account/editProfileSchema";

import {
  FormInputField,
  FormPasswordField,
} from "@/components/common/fields/rhf";
import {
  PencilSquareIcon,
  CheckIcon,
  UserCircleIcon,
} from "@heroicons/react/24/outline";
import { Button, Callout } from "@shellhub/design-system/primitives";
import PageLoader from "@/components/common/PageLoader";
import SettingsSection from "@/components/settings/SettingsSection";
import SettingsField from "@/components/settings/SettingsField";

/**
 * Edits the signed-in user's own name and emails. Changing either email asks for the current
 * password. Where the instance confirms a new email by mail (Cloud), the account keeps the old one
 * until the link sent to the new one is opened, and onEmailPending receives the new address; on
 * other editions the change applies at once. An account whose emails its identity provider
 * manages edits only its name.
 */
export function EditProfileModal({
  open,
  onClose,
  currentName,
  currentEmail,
  currentRecoveryEmail,
  emailsManagedExternally = false,
  onEmailPending,
}: {
  open: boolean;
  onClose: () => void;
  currentName: string;
  currentEmail: string;
  currentRecoveryEmail: string;
  emailsManagedExternally?: boolean;
  onEmailPending?: (email: string) => void;
}) {
  const updateProfile = useAuthStore((s) => s.updateProfile);

  const current: CurrentProfileValues = {
    name: currentName,
    email: currentEmail,
    recoveryEmail: currentRecoveryEmail,
  };
  const schema = editProfileSchema(current);

  const form = useDrawerForm(open, schema, {
    name: currentName,
    email: currentEmail,
    recoveryEmail: currentRecoveryEmail,
    currentPassword: "",
  });
  const { control, trigger, setError, clearErrors, formState } = form;

  const [watchedEmail, watchedRecoveryEmail] = useWatch({
    control,
    name: ["email", "recoveryEmail"],
  });
  const needsPassword = changesAnAddress(
    { email: watchedEmail, recoveryEmail: watchedRecoveryEmail },
    current,
  );

  const onValid = async (values: EditProfileFormValues) => {
    clearErrors("root");

    const dirty = formState.dirtyFields;
    const data: {
      name?: string;
      email?: string;
      recovery_email?: string;
      current_password?: string;
    } = {};
    if (dirty.name) data.name = values.name.trim();
    if (dirty.email) data.email = values.email;
    if (dirty.recoveryEmail) data.recovery_email = values.recoveryEmail;
    if (changesAnAddress(values, current)) {
      data.current_password = values.currentPassword;
    }

    try {
      await updateProfile(data);
      const savedEmail = useAuthStore.getState().email ?? "";
      const held =
        data.email !== undefined &&
        savedEmail.toLowerCase() !== data.email.toLowerCase();
      onClose();
      if (held && data.email) onEmailPending?.(data.email);
    } catch (err) {
      if (isSdkError(err) && err.status === 403) {
        setError("currentPassword", {
          type: "server",
          message: "Current password is incorrect.",
        });
        return;
      }

      const fields = apiErrorFields(err);
      const emailRefusal = emailRefusalMessage(fields.email);
      const recoveryEmailRefusal = emailRefusalMessage(fields.recovery_email);
      if (emailRefusal) setError("email", { type: "server", message: emailRefusal });
      if (recoveryEmailRefusal) {
        setError("recoveryEmail", { type: "server", message: recoveryEmailRefusal });
      }
      if (emailRefusal || recoveryEmailRefusal) return;

      const status = isSdkError(err) ? err.status : undefined;
      const errorMessages: Record<number, string> = {
        400: "Some fields have invalid values. Review and try again.",
        409: "That email is already in use.",
      };
      const errorMessage =
        errorMessages[status ?? 0] ?? "Failed to update profile.";

      setError("root", { message: errorMessage });
    }
  };

  return (
    <FormModal
      form={form}
      onSubmit={onValid}
      open={open}
      onClose={onClose}
      icon={<UserCircleIcon />}
      title="Edit profile"
      description="Your name and the emails ShellHub signs you in with and writes to."
      submitLabel="Save"
      requireDirty
      submitIcon={<CheckIcon className="w-4 h-4" strokeWidth={2} />}
    >
      <FormInputField
        name="name"
        control={control}
        id="profile-name"
        label="Name"
        placeholder="Your name"
        hint="1-64 characters"
        maxLength={64}
        onValueChange={() => clearErrors("root")}
      />
      <FormInputField
        name="email"
        control={control}
        id="profile-email"
        label="Email"
        type="email"
        placeholder="you@example.com"
        disabled={emailsManagedExternally}
        hint={
          emailsManagedExternally
            ? "Managed by your identity provider."
            : undefined
        }
        onValueChange={() => {
          void trigger("recoveryEmail");
          clearErrors("root");
        }}
      />
      <FormInputField
        name="recoveryEmail"
        control={control}
        id="profile-recovery-email"
        label="Recovery Email"
        type="email"
        placeholder="recovery@example.com"
        disabled={emailsManagedExternally}
        hint={
          emailsManagedExternally
            ? "Managed by your identity provider."
            : "Optional. Used for account recovery if you lose access."
        }
        onValueChange={() => clearErrors("root")}
      />
      {needsPassword && (
        <FormPasswordField
          name="currentPassword"
          control={control}
          id="profile-current-password"
          label="Current Password"
          autoComplete="current-password"
          hint="Required to change an email."
          onValueChange={() => clearErrors("root")}
        />
      )}
    </FormModal>
  );
}

/**
 * The profile section of the account: name, email and recovery email, and the modal that edits
 * them.
 */
export default function AccountProfile() {
  const { name, email, recoveryEmail, origin } = useAuthStore();
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [pendingEmail, setPendingEmail] = useState("");

  if (!name && !email) {
    return <PageLoader label="Loading profile" padding="lg" />;
  }

  return (
    <>
      <SettingsSection
        title="Profile"
        description="How you appear in ShellHub and how it reaches you."
        action={
          <Button
            size="sm"
            variant="secondary"
            onClick={() => setEditModalOpen(true)}
            icon={<PencilSquareIcon className="w-4 h-4" />}
          >
            Edit
          </Button>
        }
      >
        <SettingsField title="Name" description="Your display name.">
          <span className="text-sm font-mono text-text-secondary">{name}</span>
        </SettingsField>

        <SettingsField
          title="Email"
          description="How you sign in and where account mail goes."
        >
          <span className="text-sm font-mono text-text-secondary">{email}</span>
          {pendingEmail && (
            <Callout variant="info" className="mt-2">
              We sent a confirmation link to {pendingEmail}. Your email changes once you open it.
            </Callout>
          )}
        </SettingsField>

        <SettingsField
          title="Recovery email"
          description="Where account recovery goes if you lose access."
        >
          <span className="text-sm font-mono text-text-secondary">
            {recoveryEmail || (
              <span className="text-text-muted italic font-sans">Not set</span>
            )}
          </span>
        </SettingsField>
      </SettingsSection>
      <EditProfileModal
        open={editModalOpen}
        onClose={() => setEditModalOpen(false)}
        currentName={name ?? ""}
        currentEmail={email ?? ""}
        currentRecoveryEmail={recoveryEmail ?? ""}
        emailsManagedExternally={origin === "saml"}
        onEmailPending={setPendingEmail}
      />
    </>
  );
}
