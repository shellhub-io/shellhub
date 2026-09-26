import { useState } from "react";
import { useDrawerForm } from "@/hooks/useDrawerForm";
import { useAuthStore } from "@/stores/authStore";
import FormModal from "@/components/common/FormModal";
import { isSdkError } from "@/api/errors";
import {
  editProfileSchema,
  type EditProfileFormValues,
  type CurrentProfileValues,
} from "@/pages/account/editProfileSchema";

import {
  FormInputField,
} from "@/components/common/fields/rhf";
import {
  PencilSquareIcon,
  CheckIcon,
  UserCircleIcon,
} from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import PageLoader from "@/components/common/PageLoader";
import SettingsSection from "@/components/settings/SettingsSection";
import SettingsField from "@/components/settings/SettingsField";

/**
 * Edits the signed-in user's own name and emails. Changing the email starts a
 * re-confirmation, so the account keeps the old address until the new one is verified.
 */
export function EditProfileModal({
  open,
  onClose,
  currentName,
  currentEmail,
  currentRecoveryEmail,
}: {
  open: boolean;
  onClose: () => void;
  currentName: string;
  currentEmail: string;
  currentRecoveryEmail: string;
}) {
  const updateProfile = useAuthStore((s) => s.updateProfile);

  const current: CurrentProfileValues = {
    name: currentName,
    email: currentEmail,
  };
  const schema = editProfileSchema(current);

  const form = useDrawerForm(open, schema, {
    name: currentName,
    email: currentEmail,
    recoveryEmail: currentRecoveryEmail,
  });
  const { control, trigger, setError, clearErrors, formState } = form;

  const onValid = async (values: EditProfileFormValues) => {
    clearErrors("root");

    const dirty = formState.dirtyFields;
    const data: {
      name?: string;
      email?: string;
      recovery_email?: string;
    } = {};
    if (dirty.name) data.name = values.name;
    if (dirty.email) data.email = values.email;
    if (dirty.recoveryEmail) data.recovery_email = values.recoveryEmail;

    try {
      await updateProfile(data);
      onClose();
    } catch (err) {
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
        hint="Optional. Used for account recovery if you lose access."
        onValueChange={() => clearErrors("root")}
      />
    </FormModal>
  );
}

/**
 * The profile section of the account: name, email and recovery email, and the modal that edits
 * them.
 */
export default function AccountProfile() {
  const { name, email, recoveryEmail } = useAuthStore();
  const [editModalOpen, setEditModalOpen] = useState(false);

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
      />
    </>
  );
}
