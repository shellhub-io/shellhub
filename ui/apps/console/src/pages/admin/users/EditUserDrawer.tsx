import { useState, type FormEvent } from "react";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useUpdateUser } from "@/hooks/useAdminUserMutations";
import { useAuthStore } from "@/stores/authStore";
import { isSdkError } from "@/api/errors";
import Drawer from "@/components/common/Drawer";
import { Button } from "@shellhub/design-system/primitives";
import UserFormFields from "./UserFormFields";
import { useUserForm } from "./useUserForm";
import type { UserAdminResponse } from "@/client";

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
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button
            variant="primary"
            onClick={() => void handleSubmit()}
            disabled={!form.isSubmittable || updateUser.isPending}
            loading={updateUser.isPending}
          >
            Save Changes
          </Button>
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
