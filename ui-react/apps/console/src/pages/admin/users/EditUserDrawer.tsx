import { useState, type FormEvent } from "react";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useUpdateUser } from "@/hooks/useAdminUserMutations";
import { useAuthStore } from "@/stores/authStore";
import { isSdkError } from "@/api/errors";
import Drawer from "@/components/common/Drawer";
import UserFormFields from "./UserFormFields";
import { useUserForm } from "./useUserForm";
import type { UserAdminResponse } from "@/client";
import Spinner from "@/components/common/Spinner";

interface EditUserDrawerProps {
  open: boolean;
  onClose: () => void;
  user: UserAdminResponse | null;
}

export default function EditUserDrawer({
  open,
  onClose,
  user,
}: EditUserDrawerProps) {
  const updateUser = useUpdateUser();
  const currentUsername = useAuthStore((s) => s.username);
  const form = useUserForm({ mode: "edit" });
  const [submitError, setSubmitError] = useState("");

  useResetOnOpen(open, () => {
    form.reset(user);
    setSubmitError("");
  });

  const isSelf = user?.username === currentUsername;
  const canChangeConfirmed = user?.status !== "confirmed";
  const disableAdmin = !!user?.admin && isSelf;

  const handleSubmit = async (e?: FormEvent) => {
    e?.preventDefault();
    if (!user || !form.validateAll()) return;
    setSubmitError("");
    try {
      await updateUser.mutateAsync({
        path: { id: user.id },
        body: form.buildPayload(user),
      });
      onClose();
    } catch (err) {
      if (isSdkError(err) && err.status === 409) {
        setSubmitError("A user with this email or username already exists.");
      } else {
        setSubmitError("Failed to update user. Please try again.");
      }
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Edit User"
      subtitle={
        user ? <span className="font-mono">{user.username}</span> : undefined
      }
      footer={
        <>
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={() => void handleSubmit()}
            disabled={!form.isSubmittable || updateUser.isPending}
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
          >
            {updateUser.isPending && (
              <Spinner tone="onPrimary" />
            )}
            Save Changes
          </button>
        </>
      }
    >
      <form
        onSubmit={(e) => void handleSubmit(e)}
        className="space-y-5"
        noValidate
      >
        <UserFormFields
          form={form}
          idPrefix="edit-user"
          autoFocus={open}
          canChangeConfirmed={canChangeConfirmed}
          disableAdmin={disableAdmin}
        />
        {submitError && (
          <p role="alert" className="text-2xs text-accent-red">
            {submitError}
          </p>
        )}
      </form>
    </Drawer>
  );
}
